param(
    [string]$Project = '.', [string]$Version = '', [string]$Dest = '',
    [string]$LocalBin = '', [string]$Sha256 = '', [switch]$Force,
    [string]$ManagerVersion = '', [string]$LocalManagerBin = '', [string]$ManagerSha256 = '',
    [switch]$DryRun, [switch]$InitConfig,
    [ValidateSet('', 'codex', 'opencode')][string]$CaptureSource = '',
    [string]$CaptureStore = '', [string]$CaptureStateDir = ''
)
# Bootstrap only. Native manager owns configuration and executable installation.
$ErrorActionPreference = 'Stop'
$stage = ''; $scaffold = ''; $scaffoldContent = ''
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
function Run-Probe([string]$Executable, [string[]]$Arguments, [int]$TimeoutMS = 10000) {
    $quoted = [Collections.Generic.List[string]]::new()
    foreach ($argument in $Arguments) {
        $encoded = [regex]::Replace($argument, '(\\*)"', '$1$1\"')
        $encoded = [regex]::Replace($encoded, '(\\+)$', '$1$1')
        $quoted.Add('"' + $encoded + '"')
    }
    $start = [Diagnostics.ProcessStartInfo]::new()
    if ((Get-Location).Provider.Name -cne 'FileSystem') { throw 'Native setup requires a filesystem working directory' }
    $start.WorkingDirectory = (Get-Location).ProviderPath
    $start.FileName = $Executable; $start.Arguments = [string]::Join(' ', $quoted)
    $start.UseShellExecute = $false; $start.RedirectStandardOutput = $true; $start.RedirectStandardError = $true
    $start.StandardOutputEncoding = [Text.UTF8Encoding]::new($false)
    $process = [Diagnostics.Process]::Start($start)
    try {
        $result = Read-ProbeResult $process $TimeoutMS 1048576 $true
        if ($result.Code -ne 0) { throw "Native setup probe failed: $($result.Diagnostic)$($result.Output)" }
        if ($result.Diagnostic) {
            $bytes = [Text.Encoding]::UTF8.GetBytes($result.Diagnostic)
            $stream = [Console]::OpenStandardError(); $stream.Write($bytes, 0, $bytes.Length); $stream.Flush()
        }
        return $result.Output.TrimEnd("`r", "`n")
    } finally { $process.Dispose() }
}
function Exact-Version([string]$Value) {
    return $Value.Length -le 128 -and $Value -cmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$'
}
try {
    if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) { throw 'Use download.sh on POSIX' }
    if ([Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString() -cne 'X64') { throw 'Windows native artifacts support amd64' }
    if ($env:AGENTIC_LENS_BIN_DIR) { throw 'Use local agenticlens.install_root and -Dest' }
    $projectPath = Resolve-PhysicalDirectory $Project $true
    if (-not $Dest) { $Dest = [IO.Path]::Combine([Environment]::GetFolderPath('UserProfile'), '.agenticlens') }
    $Dest = Resolve-PhysicalDirectory $Dest
    $entrypoint = [IO.Path]::Combine($Dest, 'agentic-lens.exe')
    if ($CaptureSource -and -not [IO.Path]::IsPathRooted($CaptureStore)) { throw 'Capture requires an explicit absolute source store' }
    if (-not $CaptureSource -and ($CaptureStore -or $CaptureStateDir)) { throw 'Capture options require -CaptureSource' }
    $stage = [IO.Path]::Combine([IO.Path]::GetTempPath(), 'agenticlens-bootstrap-' + [guid]::NewGuid().ToString('N'))
    [IO.Directory]::CreateDirectory($stage) | Out-Null
    $manager = [IO.Path]::Combine($stage, 'manager.exe')
    if ($LocalManagerBin) {
        if ($ManagerSha256 -notmatch '^[a-fA-F0-9]{64}$') { throw 'Local manager requires explicit checksum' }
        [IO.File]::Copy((Resolve-Path -LiteralPath $LocalManagerBin).ProviderPath, $manager)
        if ((Hash-File $manager) -cne $ManagerSha256.ToLowerInvariant()) { throw 'Local manager checksum mismatch' }
    } elseif ([IO.File]::Exists($entrypoint)) {
        if ((Run-Probe $entrypoint @('manager', 'protocol')) -cne '1') { throw 'Replace the older installation explicitly; no wrapper compatibility path is installed' }
        [IO.File]::Copy($entrypoint, $manager)
    } else {
        if (-not $ManagerVersion) { $ManagerVersion = $Version }
        if (-not (Exact-Version $ManagerVersion)) { throw 'First install requires exact -ManagerVersion or verified -LocalManagerBin' }
        if ($DryRun) { Write-SetupLine ("Manager release: $ManagerVersion; native configuration validation pending bootstrap; no network or destination writes."); exit 0 }
        $repository = 'NiltonDuarte/AgenticLens'; if ($env:AGENTIC_LENS_REPO) { $repository = $env:AGENTIC_LENS_REPO }
        $base = "https://github.com/$repository/releases/download/$ManagerVersion"
        $checksums = (Invoke-WebRequest -UseBasicParsing -Uri "$base/SHA256SUMS").Content
        Invoke-WebRequest -UseBasicParsing -Uri "$base/agentic-lens-windows-amd64.exe" -OutFile $manager
        $matching = @($checksums -split '\r?\n' | Where-Object { $_ -match '^([a-fA-F0-9]{64})\s+\*?(?:bin/)?agentic-lens-windows-amd64\.exe$' })
        if ($matching.Count -ne 1 -or (Hash-File $manager) -cne ($matching[0] -split '\s+')[0].ToLowerInvariant()) { throw 'Missing, ambiguous or mismatching manager release checksum' }
    }
    $stream = [IO.File]::OpenRead($manager)
    try { if ($stream.ReadByte() -ne 77 -or $stream.ReadByte() -ne 90) { throw 'Bootstrap is not a Windows native executable' } } finally { $stream.Dispose() }
    if ((Run-Probe $manager @('manager', 'protocol')) -cne '1') { throw 'Native manager interface unavailable' }
    $actualManagerVersion = Run-Probe $manager @('--manager-version', '--short')
    if (-not (Exact-Version $actualManagerVersion) -or ($ManagerVersion -and $ManagerVersion -cne $actualManagerVersion)) { throw 'Manager version mismatch' }
    $ManagerVersion = $actualManagerVersion
    $resolveProject = $projectPath
    if ($InitConfig) {
        if (-not (Exact-Version $Version)) { throw '-InitConfig requires exact -Version' }
        foreach ($suffix in @('yaml', 'yml')) { if (Test-Path -LiteralPath ([IO.Path]::Combine($projectPath, ".agenticlens-config.$suffix"))) { throw 'Shared configuration exists; edit its pin explicitly' } }
        $scaffoldContent = "# AgenticLens bootstrap scaffold v1: incomplete project configuration`nagenticlens:`n  version: $Version`n"
        [IO.File]::WriteAllText([IO.Path]::Combine($stage, '.agenticlens-config.yaml'), $scaffoldContent, [Text.UTF8Encoding]::new($false))
        $resolveProject = $stage
    }
    $context = (Run-Probe $manager @('--al-project', $resolveProject, '--al-resolve-cache')) -split '\r?\n'
    if ($context.Count -ne 4) { throw 'Native configuration protocol incompatible' }
    $pin = $context[0]
    if ($Version -and $Version -cne $pin) { throw '-Version conflicts with shared YAML pin' }
    if ($env:AGENTIC_LENS_VERSION -and $env:AGENTIC_LENS_VERSION -cne $pin) { throw 'AGENTIC_LENS_VERSION conflicts with shared YAML pin' }
    if (-not $InitConfig) { $projectPath = $context[1] }
    if ($DryRun) { Write-SetupLine ("Project: $projectPath`nEngine version: $pin`nManager version: $ManagerVersion`nNative command: $entrypoint`nNo network or destination writes. Required PATH entry: $Dest"); exit 0 }
    if ($InitConfig) {
        $scaffold = [IO.Path]::Combine($projectPath, '.agenticlens-config.yaml')
        $stream = [IO.File]::Open($scaffold, [IO.FileMode]::CreateNew)
        try { $bytes = [Text.Encoding]::UTF8.GetBytes($scaffoldContent); $stream.Write($bytes, 0, $bytes.Length) } finally { $stream.Dispose() }
    }
    $arguments = @('--al-project', $projectPath, 'install', $pin)
    if ($Force) { $arguments += '--force' }
    if ($LocalBin) { $arguments += @('--local-bin', (Resolve-Path -LiteralPath $LocalBin).ProviderPath, '--sha256', $Sha256) }
    Write-SetupLine (Run-Probe $manager $arguments 300000)
    Write-SetupLine (Run-Probe $manager @('self-update', '--version', $ManagerVersion, '--local-bin', $manager, '--sha256', (Hash-File $manager), '--dest', $entrypoint) 300000)
    if ($InitConfig) {
        Write-SetupLine 'Configuration incomplete: review and approve explicit shared/local YAML candidates before capture setup.'
        $quotedEntrypoint = $entrypoint.Replace("'", "''")
        $quotedProject = $projectPath.Replace("'", "''")
        $quotedOutput = ([IO.Path]::Combine($projectPath, '.agenticlens-config.yaml')).Replace("'", "''")
        $quotedPin = $pin.Replace("'", "''")
        Write-SetupLine ("Complete unchanged owned scaffold: & '$quotedEntrypoint' --al-project '$quotedProject' init --complete-scaffold --agenticlens-version '$quotedPin' --output '$quotedOutput' --settings 'APPROVED_SHARED.yaml' --local-settings 'APPROVED_LOCAL.yaml' --workspaces 'DIR'")
    }
    if ($CaptureSource) {
        if ($InitConfig) { throw 'Complete project configuration before capture installation' }
        $arguments = @('--al-project', $projectPath, 'capture', 'install', '--source', $CaptureSource, '--store', $CaptureStore, '--repo', $projectPath, '--wrapper', $entrypoint)
        if ($CaptureSource -ceq 'codex') { $arguments = @('--al-project', $projectPath, 'capture', 'install-user', '--repo', $projectPath, '--entrypoint', $entrypoint, '--store', $CaptureStore) }
        if ($CaptureStateDir) { $arguments += @('--state-dir', $CaptureStateDir) }
        Write-SetupLine (Run-Probe $entrypoint $arguments 300000)
    }
    Write-SetupLine ("Installed native manager: $entrypoint`nEngine version: $pin`nRequired PATH entry: $Dest`nNo profile changed. Runtime trust, binding and source extraction require verification.")
} catch {
    if ($scaffold -and [IO.File]::Exists($scaffold) -and [IO.File]::ReadAllText($scaffold) -ceq $scaffoldContent) { [IO.File]::Delete($scaffold) }
    Write-SetupError $_.Exception.Message
    exit 1
} finally {
    if ($stage -and [IO.Directory]::Exists($stage)) {
        # File.Copy may retain ReadOnly on the two files this bootstrap owns.
        # Never change attributes on user inputs, installed engines or cache files.
        foreach ($name in @('manager.exe', '.agenticlens-config.yaml')) {
            $ownedStageFile = [IO.Path]::Combine($stage, $name)
            if ([IO.File]::Exists($ownedStageFile)) {
                $attributes = [IO.File]::GetAttributes($ownedStageFile)
                if (($attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Owned bootstrap staging file became a reparse point; cleanup refused' }
                [IO.File]::SetAttributes($ownedStageFile, [IO.FileAttributes]([int]$attributes -band (-bnot [int][IO.FileAttributes]::ReadOnly)))
            }
        }
        [IO.Directory]::Delete($stage, $true)
    }
}
