param(
    [string]$Project = '.', [string]$Version = '', [string]$Dest = '',
    [string]$LocalBin = '', [string]$Sha256 = '', [switch]$Force,
    [switch]$DryRun, [switch]$InitConfig,
    [ValidateSet('', 'codex', 'opencode')][string]$CaptureSource = '',
    [string]$CaptureStore = '', [string]$CaptureStateDir = ''
)
# Explicit native Windows installer. No policy, trust or persistent PATH changes.
$ErrorActionPreference = 'Stop'
$targets = [Collections.Generic.List[string]]::new()
$backups = [Collections.Generic.List[string]]::new()
$stage = ''; $cacheLock = ''; $wrapperLock = ''; $keepStage = $false; $captureApplyStarted = $false
# Resolve reparse targets before following later .. components, including
# not-yet-created cache suffixes. Never use GetFullPath to collapse them first.
function Resolve-PhysicalDirectory([string]$Path, [bool]$RequireExisting = $false) {
    $raw = $Path.Replace('/', '\')
    if ($raw -notmatch '^([A-Za-z]:\\|\\\\[^\\]+\\[^\\]+)') {
        if ($raw.StartsWith('\') -or $raw -match '^[A-Za-z]:') { throw 'Use an absolute drive/UNC path or an ordinary relative directory' }
        $raw = (Get-Location).ProviderPath.TrimEnd('\') + '\' + $raw
    }
    $volume = [IO.Path]::GetPathRoot($raw)
    $current = $volume
    $pending = [Collections.Generic.List[string]]::new()
    foreach ($part in ($raw.Substring($volume.Length) -split '\\')) { $pending.Add($part) }
    $links = 0
    while ($pending.Count -gt 0) {
        $part = $pending[0]; $pending.RemoveAt(0)
        if (-not $part -or $part -ceq '.') { continue }
        if ($part -ceq '..') {
            $parent = [IO.Directory]::GetParent($current)
            if ($null -ne $parent) { $current = $parent.FullName }
            continue
        }
        $candidate = [IO.Path]::Combine($current, $part)
        $item = Get-Item -LiteralPath $candidate -Force -ErrorAction SilentlyContinue
        if ($null -ne $item -and ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
            $links++; if ($links -gt 255) { throw 'Excessive reparse-point traversal' }
            $target = [string]@($item.Target)[0]
            if (-not $target) { throw 'Unsupported reparse-point target; use a physical installation/project path' }
            if ($target.StartsWith('\??\')) { $target = $target.Substring(4) }
            if ($target.StartsWith('\\?\')) { $target = $target.Substring(4) }
            if ($target -match '^([A-Za-z]:\\|\\\\[^\\]+\\[^\\]+)') {
                $current = [IO.Path]::GetPathRoot($target)
                $target = $target.Substring($current.Length)
            }
            $pending.InsertRange(0, [string[]]($target -split '\\'))
            continue
        }
        if ($null -ne $item -and -not $item.PSIsContainer) { throw 'Directory path crosses a non-directory' }
        $current = $candidate
    }
    if ($RequireExisting -and -not [IO.Directory]::Exists($current)) { throw 'Project directory does not exist' }
    return $current.TrimEnd('\') + $(if ($current -eq $volume) { '\' } else { '' })
}
function Write-SetupLine([string]$Text) {
    $bytes = [Text.Encoding]::UTF8.GetBytes($Text + "`n")
    $stream = [Console]::OpenStandardOutput(); $stream.Write($bytes, 0, $bytes.Length); $stream.Flush()
}
function Write-SetupError([string]$Text) {
    $bytes = [Text.Encoding]::UTF8.GetBytes($Text + "`n")
    $stream = [Console]::OpenStandardError(); $stream.Write($bytes, 0, $bytes.Length); $stream.Flush()
}
function Hash-File([string]$Path) {
    $stream = $null; $sha256 = $null
    try {
        # Preserve literal PowerShell paths, including its current directory.
        $literalPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Path)
        $stream = [IO.File]::OpenRead($literalPath)
        $sha256 = [Security.Cryptography.SHA256]::Create()
        return [BitConverter]::ToString($sha256.ComputeHash($stream)).Replace('-', '').ToLowerInvariant()
    } finally {
        if ($null -ne $sha256) { $sha256.Dispose() }
        if ($null -ne $stream) { $stream.Dispose() }
    }
}
# Read both redirected streams concurrently; neither child pipe can fill while
# waiting for the other or for exit. Bounds apply to bytes before UTF-8 decoding.
function Read-ProbeResult([Diagnostics.Process]$Process, [int]$TimeoutMS, [int]$Limit, [bool]$CleanupTree) {
    $outBuffer = [byte[]]::new(8192); $errBuffer = [byte[]]::new(8192)
    $output = [IO.MemoryStream]::new(); $diagnostic = [IO.MemoryStream]::new()
    $outDone = $false; $errDone = $false
    $outRead = $Process.StandardOutput.BaseStream.ReadAsync($outBuffer, 0, $outBuffer.Length)
    $errRead = $Process.StandardError.BaseStream.ReadAsync($errBuffer, 0, $errBuffer.Length)
    $clock = [Diagnostics.Stopwatch]::StartNew()
    try {
        while ($true) {
            if (-not $outDone -and $outRead.IsCompleted) {
                $count = $outRead.GetAwaiter().GetResult()
                if (-not $count) { $outDone = $true }
                else {
                    if ($output.Length + $count -gt $Limit) { throw 'Native setup stdout exceeded its byte limit' }
                    $output.Write($outBuffer, 0, $count)
                    $outRead = $Process.StandardOutput.BaseStream.ReadAsync($outBuffer, 0, $outBuffer.Length)
                }
            }
            if (-not $errDone -and $errRead.IsCompleted) {
                $count = $errRead.GetAwaiter().GetResult()
                if (-not $count) { $errDone = $true }
                else {
                    if ($diagnostic.Length + $count -gt $Limit) { throw 'Native setup stderr exceeded its byte limit' }
                    $diagnostic.Write($errBuffer, 0, $count)
                    $errRead = $Process.StandardError.BaseStream.ReadAsync($errBuffer, 0, $errBuffer.Length)
                }
            }
            if ($outDone -and $errDone -and $Process.HasExited) { break }
            if ($clock.ElapsedMilliseconds -ge $TimeoutMS) { throw 'Native setup probe timed out' }
            $pending = [Collections.Generic.List[Threading.Tasks.Task]]::new()
            if (-not $outDone) { $pending.Add($outRead) }
            if (-not $errDone) { $pending.Add($errRead) }
            if ($pending.Count) { [void][Threading.Tasks.Task]::WaitAny($pending.ToArray(), 10) }
            else { [void]$Process.WaitForExit(10) }
        }
        $utf8 = [Text.UTF8Encoding]::new($false, $true)
        return @{ Code = $Process.ExitCode; Output = $utf8.GetString($output.ToArray()); Diagnostic = $utf8.GetString($diagnostic.ToArray()) }
    } catch {
        if ($CleanupTree) { Stop-OwnedProbe $Process }
        elseif (-not $Process.HasExited) { $Process.Kill(); [void]$Process.WaitForExit(500) }
        throw
    } finally {
        $Process.StandardOutput.Dispose(); $Process.StandardError.Dispose()
        $output.Dispose(); $diagnostic.Dispose()
    }
}
function Stop-OwnedProbe([Diagnostics.Process]$Process) {
    if ($Process.HasExited) { return }
    $start = [Diagnostics.ProcessStartInfo]::new()
    $start.FileName = [IO.Path]::Combine($env:SystemRoot, 'System32', 'taskkill.exe')
    $start.Arguments = '/PID ' + [string]$Process.Id + ' /T /F'
    $start.UseShellExecute = $false; $start.CreateNoWindow = $true
    $start.RedirectStandardOutput = $true; $start.RedirectStandardError = $true
    $cleanup = $null
    try {
        $cleanup = [Diagnostics.Process]::Start($start)
        $result = Read-ProbeResult $cleanup 1000 16384 $false
        if ($result.Code -ne 0 -and -not $Process.HasExited) { throw 'Owned process-tree cleanup failed' }
        [void]$Process.WaitForExit(500)
    } catch {
        if (-not $Process.HasExited) { $Process.Kill(); [void]$Process.WaitForExit(500) }
        Write-SetupError ('Owned setup process-tree cleanup incomplete; inspect retained setup state and running capture processes.')
    } finally { if ($null -ne $cleanup) { $cleanup.Dispose() } }
}
function Run-Probe([string]$Executable, [string[]]$Arguments) {
    $quoted = [Collections.Generic.List[string]]::new()
    foreach ($argument in $Arguments) {
        $encoded = [regex]::Replace($argument, '(\\*)"', '$1$1\"')
        $encoded = [regex]::Replace($encoded, '(\\+)$', '$1$1')
        $quoted.Add('"' + $encoded + '"')
    }
    $start = [Diagnostics.ProcessStartInfo]::new()
    $start.FileName = $Executable; $start.Arguments = [string]::Join(' ', $quoted)
    $start.UseShellExecute = $false; $start.RedirectStandardOutput = $true; $start.RedirectStandardError = $true
    $start.StandardOutputEncoding = [Text.UTF8Encoding]::new($false)
    $process = [Diagnostics.Process]::Start($start)
    try {
        $result = Read-ProbeResult $process 10000 1048576 $true
        if ($result.Code -ne 0) { throw "Native setup probe failed: $($result.Diagnostic)$($result.Output)" }
        if ($result.Diagnostic) {
            $bytes = [Text.Encoding]::UTF8.GetBytes($result.Diagnostic)
            $stream = [Console]::OpenStandardError(); $stream.Write($bytes, 0, $bytes.Length); $stream.Flush()
        }
        return $result.Output.TrimEnd("`r", "`n")
    } finally { $process.Dispose() }
}
function Activate([string]$Candidate, [string]$Target) {
    [IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($Target)) | Out-Null
    $backup = ''
    if ([IO.File]::Exists($Target)) { $backup = [IO.Path]::Combine($stage, [guid]::NewGuid().ToString('N') + '.previous'); [IO.File]::Copy($Target, $backup) }
    $targets.Add($Target); $backups.Add($backup)
    $near = $Target + '.candidate.' + [guid]::NewGuid().ToString('N')
    try {
        [IO.File]::Copy($Candidate, $near)
        if ([IO.File]::Exists($Target)) { [IO.File]::Replace($near, $Target, [NullString]::Value) } else { [IO.File]::Move($near, $Target) }
    } finally { if ([IO.File]::Exists($near)) { [IO.File]::Delete($near) } }
}
try {
    if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) { throw 'Use download.sh on the executing POSIX environment' }
    if ([Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString() -cne 'X64') { throw 'Windows arm64/x86 native release artifacts are unavailable' }
    if ($env:AGENTIC_LENS_BIN_DIR) { throw 'AGENTIC_LENS_BIN_DIR retired; use local agenticlens.install_root and -Dest' }
    $projectPath = Resolve-PhysicalDirectory $Project $true
    $source = [IO.Path]::Combine($PSScriptRoot, 'agentic-lens-wrapper.ps1')
    $shell = [IO.Path]::Combine($env:SystemRoot, 'System32', 'WindowsPowerShell', 'v1.0', 'powershell.exe')
    if (-not $Dest) { $Dest = [IO.Path]::Combine([Environment]::GetFolderPath('UserProfile'), '.agenticlens') }
    $Dest = Resolve-PhysicalDirectory $Dest
    if ($CaptureSource -and -not [IO.Path]::IsPathRooted($CaptureStore)) { throw 'Capture requires an explicit absolute source store' }
    if (-not $CaptureSource -and ($CaptureStore -or $CaptureStateDir)) { throw 'Capture options require -CaptureSource' }
    if ($InitConfig) {
        if ($Version.Length -gt 128 -or $Version -cnotmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$') { throw '-InitConfig requires an exact -Version release' }
        foreach ($suffix in @('yaml', 'yml', 'json')) { if (Test-Path -LiteralPath ([IO.Path]::Combine($projectPath, ".agenticlens-config.$suffix"))) { throw 'Shared configuration exists; edit its pin explicitly' } }
        if ($DryRun) { Write-SetupLine ("Would create shared pin $Version and explicitly install Windows amd64; runtime verification pending."); exit 0 }
        $pin = [IO.Path]::Combine($projectPath, '.agenticlens-config.yaml')
        $stream = [IO.File]::Open($pin, [IO.FileMode]::CreateNew)
        try { $bytes = [Text.Encoding]::UTF8.GetBytes("# AgenticLens bootstrap scaffold v1: incomplete project configuration`nagenticlens:`n  version: $Version`n"); $stream.Write($bytes, 0, $bytes.Length) } finally { $stream.Dispose() }
        $targets.Add($pin); $backups.Add('')
    }
    $context = (Run-Probe $shell @('-NoLogo', '-NoProfile', '-NonInteractive', '-File', $source, '--al-project', $projectPath, '--al-resolve-cache')) -split '\r?\n'
    if ($context.Count -ne 4) { throw 'Common wrapper resolution protocol incompatible' }
    $pinVersion = $context[0]; $projectPath = $context[1]; $installRoot = $context[2]; $native = $context[3]
    if ($Version -and $Version -cne $pinVersion) { throw '-Version conflicts with shared YAML pin' }
    $wrapper = [IO.Path]::Combine($Dest, 'agentic-lens.ps1')
    if ($DryRun) { Write-SetupLine ("Project: $projectPath`nVersion: $pinVersion`nWrapper: $wrapper`nNative: $native`nRequired PATH entry: $Dest; no profile changed. Runtime verification pending."); exit 0 }
    [IO.Directory]::CreateDirectory($installRoot) | Out-Null; [IO.Directory]::CreateDirectory($Dest) | Out-Null
    $candidateLock = [IO.Path]::Combine($installRoot, ".install-$pinVersion.lock")
    # Atomic file creation excludes concurrent installers without stale lock removal.
    $lockFile = [IO.File]::Open($candidateLock, [IO.FileMode]::CreateNew); $lockFile.Dispose(); $cacheLock = $candidateLock
    $candidateLock = [IO.Path]::Combine($Dest, '.agenticlens-wrapper-install.lock')
    $lockFile = [IO.File]::Open($candidateLock, [IO.FileMode]::CreateNew); $lockFile.Dispose(); $wrapperLock = $candidateLock
    $stage = [IO.Path]::Combine($installRoot, '.install-stage.' + [guid]::NewGuid().ToString('N'))
    [IO.Directory]::CreateDirectory($stage) | Out-Null
    if ([IO.File]::Exists($wrapper) -and (Hash-File $wrapper) -cne (Hash-File $source)) {
        if (-not [IO.File]::Exists($wrapper + '.sha256') -or [IO.File]::ReadAllText($wrapper + '.sha256').Trim() -cne (Hash-File $wrapper)) { throw "User-edited/unowned wrapper preserved: $wrapper" }
    }
    $candidate = [IO.Path]::Combine($stage, 'agentic-lens-windows-amd64.exe')
    if ($LocalBin) {
        if ($Sha256 -notmatch '^[a-fA-F0-9]{64}$' -or (Hash-File $LocalBin) -cne $Sha256.ToLowerInvariant()) { throw '-LocalBin requires the matching explicit expected -Sha256' }
        [IO.File]::Copy((Resolve-Path -LiteralPath $LocalBin).ProviderPath, $candidate)
    } else {
        $repository = 'NiltonDuarte/AgenticLens'; if ($env:AGENTIC_LENS_REPO) { $repository = $env:AGENTIC_LENS_REPO }
        $base = "https://github.com/$repository/releases/download/$pinVersion"
        Invoke-WebRequest -UseBasicParsing -Uri "$base/agentic-lens-windows-amd64.exe" -OutFile $candidate
        $checksums = (Invoke-WebRequest -UseBasicParsing -Uri "$base/SHA256SUMS").Content
        $matching = @($checksums -split '\r?\n' | Where-Object { $_ -match '^([a-fA-F0-9]{64})\s+\*?(?:bin/)?agentic-lens-windows-amd64\.exe$' })
        if ($matching.Count -ne 1 -or (Hash-File $candidate) -cne ($matching[0] -split '\s+')[0].ToLowerInvariant()) { throw 'Missing, ambiguous or mismatching release checksum' }
    }
    $stream = [IO.File]::OpenRead($candidate)
    try { if ($stream.ReadByte() -ne 77 -or $stream.ReadByte() -ne 90) { throw 'Candidate is not a native Windows executable' } } finally { $stream.Dispose() }
    if ((Run-Probe $candidate @('version', '--short')) -cne $pinVersion) { throw 'Native version differs from shared pin' }
    $required = ''; if ($CaptureSource) { $required = '3' }
    if ([IO.File]::Exists($native)) {
        $existingStream = [IO.File]::OpenRead($native)
        try { $existingNative = $existingStream.ReadByte() -eq 77 -and $existingStream.ReadByte() -eq 90 } finally { $existingStream.Dispose() }
        if ($existingNative) { try { $existingProtocol = Run-Probe $native @('capture', 'protocol'); if ($existingProtocol -ceq '3') { $required = '3' } elseif ($existingProtocol -ceq '2' -and -not $required) { $required = '2' } } catch { } }
    }
    $observedProtocol = ''; try { $observedProtocol = Run-Probe $candidate @('capture', 'protocol') } catch { }
    if ($required -and $observedProtocol -cne $required) { throw "Native hook protocol incompatible (required $required)" }
    if ([IO.File]::Exists($native) -and (Hash-File $native) -cne (Hash-File $candidate) -and -not $Force) { throw 'Immutable cache leaf exists; -Force explicitly repairs only this version' }
    # Identical leaves retain the exact file object, including explicit repair.
    if (-not [IO.File]::Exists($native) -or (Hash-File $native) -cne (Hash-File $candidate)) { Activate $candidate $native }
    Activate $source $wrapper
    $hashFile = [IO.Path]::Combine($stage, 'wrapper.sha256'); [IO.File]::WriteAllText($hashFile, (Hash-File $source))
    Activate $hashFile ($wrapper + '.sha256')
    $dispatcherProtocol = ''
    try { $dispatcherProtocol = (Run-Probe $native @('dispatch', 'protocol')).Trim() } catch { $dispatcherProtocol = '' }
    if ($dispatcherProtocol -ceq '1') {
        # The native transaction owns bytes, checksum and machine receipt.
        $dispatcher = [IO.Path]::Combine($installRoot, 'dispatcher.exe')
        Write-SetupLine ((Run-Probe $native @('capture', 'install-dispatcher', '--repo', $projectPath, '--binary', $native)))
        $pointer = [IO.Path]::Combine($stage, 'dispatcher.pointer')
        [IO.File]::WriteAllText($pointer, $dispatcher); Activate $pointer ($wrapper + '.dispatcher')
    }

    if ($InitConfig) {
        Write-SetupLine ('Configuration incomplete: scaffold defines no project/ticket/code roots.')
        if ($observedProtocol -ceq '3') {
            $wrapperLiteral = "'" + $wrapper.Replace("'", "''") + "'"
            $projectLiteral = "'" + $projectPath.Replace("'", "''") + "'"
            Write-SetupLine ("Complete unchanged owned scaffold: & { Push-Location -LiteralPath $projectLiteral; try { powershell.exe -NoLogo -NoProfile -NonInteractive -File $wrapperLiteral --al-project $projectLiteral init --complete-scaffold --agenticlens-version $pinVersion --project-name NAME --src-folders DIRS --ticket-prefixes PREFIXES } finally { Pop-Location } }")
        }
        else { Write-SetupLine ('Selected legacy release lacks guarded init/capture capability. Ordinary CLI installed; review an exact supported release pin before adopting new setup/capture.') }
    }
    if ($CaptureSource) {
        if ($InitConfig) { throw 'Complete project configuration before enabling capture' }
        $captureArgs = @('capture', 'install', '--repo', $projectPath, '--source', $CaptureSource, '--store', $CaptureStore, '--binary', $native, '--wrapper', $wrapper)
        if ($CaptureSource -ceq 'codex') { $captureArgs = @('capture', 'install-user', '--repo', $projectPath, '--binary', $native) }
        if ($CaptureStateDir) { $captureArgs += @('--state-dir', $CaptureStateDir) }
        $captureApplyStarted = $true
        Write-SetupLine ((Run-Probe $native $captureArgs))
    }
    if ($observedProtocol -cne '3') { Write-SetupLine ("Selected legacy release $pinVersion`: hook protocol 3 and guarded init unavailable; ordinary CLI installed. New capture setup requires a reviewed exact supported pin.") }
    Write-SetupLine ("Installed wrapper: $wrapper`nNative: $native`nRequired PATH entry: $Dest`nNo profile changed. Runtime trust, binding and source extraction require verification.")
} catch {
    $failureMessage = $_.Exception.Message
    if ($captureApplyStarted) {
        # Native apply can commit before result encoding/transport fails. Keep
        # validated handlers so any committed hooks remain executable.
        $keepStage = $true
        Write-SetupError ("Capture apply incomplete; validated native/wrapper retained. Prior backups: $stage. Inspect capture status and preview owned capture install before recovery; no complete rollback is claimed.")
    }
    for ($i = $(if ($captureApplyStarted) { -1 } else { $targets.Count - 1 }); $i -ge 0; $i--) {
        try {
            if ($backups[$i]) { [IO.File]::Copy($backups[$i], $targets[$i], $true) }
            elseif ([IO.File]::Exists($targets[$i])) { [IO.File]::Delete($targets[$i]) }
        } catch { Write-SetupError ("Restore prior installation from $($backups[$i]): $($_.Exception.Message)"); $keepStage = $true }
    }
    Write-SetupError ($failureMessage)
    exit 1
} finally {
    if ($stage -and -not $keepStage -and [IO.Directory]::Exists($stage)) { [IO.Directory]::Delete($stage, $true) }
    if ($cacheLock -and [IO.File]::Exists($cacheLock)) { [IO.File]::Delete($cacheLock) }
    if ($wrapperLock -and [IO.File]::Exists($wrapperLock)) { [IO.File]::Delete($wrapperLock) }
}
