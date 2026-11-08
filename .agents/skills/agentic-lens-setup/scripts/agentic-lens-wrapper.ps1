# AgenticLens common entrypoint v1: offline shared-YAML release selection.
$ErrorActionPreference = 'Stop'
function Invoke-NativeTransport([string]$native, [string[]]$arguments, [bool]$advisory) {
    # ProcessStartInfo with inherited handles avoids PowerShell 5.1's lossy
    # empty-argument conversion and native pipeline re-encoding. Windows CRT
    # argument quoting doubles trailing backslashes and those before quotes.
    $quotedArguments = [Collections.Generic.List[string]]::new()
    foreach ($argument in $arguments) {
        $encoded = [Text.RegularExpressions.Regex]::Replace($argument, '(\\*)"', '$1$1\"')
        $encoded = [Text.RegularExpressions.Regex]::Replace($encoded, '(\\+)$', '$1$1')
        $quotedArguments.Add('"' + $encoded + '"')
    }
    $start = [Diagnostics.ProcessStartInfo]::new()
    $start.FileName = $native
    $start.UseShellExecute = $false
    $start.Arguments = [string]::Join(' ', $quotedArguments)
    # PowerShell's filesystem location can differ from the process directory.
    $location = Get-Location
    if ($location.Provider.Name -cne 'FileSystem') { throw 'Native entrypoint requires a filesystem current directory' }
    $start.WorkingDirectory = $location.ProviderPath
    $process = [Diagnostics.Process]::Start($start)
    if ($advisory) {
        if (-not $process.WaitForExit(1500)) {
            $terminate = [Diagnostics.ProcessStartInfo]::new()
            $terminate.FileName = [IO.Path]::Combine($env:SystemRoot, 'System32', 'taskkill.exe')
            $terminate.Arguments = '/PID ' + $process.Id + ' /T /F'
            $terminate.UseShellExecute = $false; $terminate.CreateNoWindow = $true
            $cleanupDeadline = [DateTime]::UtcNow.AddMilliseconds(500)
            $stop = [Diagnostics.Process]::Start($terminate)
            if (-not $stop.WaitForExit(250)) { $stop.Kill() }; $stop.Dispose()
            if (-not $process.HasExited) { $process.Kill() }
            $remaining = [Math]::Max(0, [int]($cleanupDeadline - [DateTime]::UtcNow).TotalMilliseconds)
            [void]$process.WaitForExit($remaining); $process.Dispose()
            throw 'Offline advisory check timed out; inspect capture diagnostics explicitly'
        }
    } else { $process.WaitForExit() }
    $exitCode = $process.ExitCode
    $process.Dispose()
    return $exitCode
}
function Test-OfflineAdvisory([string[]]$values) {
    $offline = $false; $index = 0
    while ($index -lt $values.Count) {
        if ($values[$index] -ceq '--al-offline') { $offline = $true; $index++; continue }
        if ($values[$index] -ceq '--al-project') { $index += 2; continue }
        break
    }
    return ($offline -and $index + 1 -lt $values.Count -and $values[$index] -ceq 'capture' -and $values[$index + 1] -ceq 'check')
}
# Native selection owns shared/local configuration policy after explicit install.
$dispatcherPath = [IO.Path]::Combine([Environment]::GetFolderPath('UserProfile'), '.agenticlens', 'dispatcher.exe')
$dispatcherPointer = $PSCommandPath + '.dispatcher'
if ([IO.File]::Exists($dispatcherPointer)) { $dispatcherPath = [IO.File]::ReadAllText($dispatcherPointer).TrimEnd("`r", "`n") }
if ([IO.File]::Exists($dispatcherPath)) {
    try { exit (Invoke-NativeTransport $dispatcherPath (@('dispatch') + $args) (Test-OfflineAdvisory $args)) }
    catch {
        $message = [Text.Encoding]::UTF8.GetBytes($_.Exception.Message + "`n")
        $errorStream = [Console]::OpenStandardError(); $errorStream.Write($message, 0, $message.Length); $errorStream.Flush()
        exit 2
    }
}
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
function Read-Bootstrap([string]$Path, [bool]$Local) {
    if ((Get-Item -LiteralPath $Path).Length -gt 1048576) { throw 'Bootstrap configuration exceeds 1 MiB' }
    $roots = [Collections.Generic.Dictionary[string,bool]]::new([StringComparer]::Ordinal)
    $children = [Collections.Generic.Dictionary[string,bool]]::new([StringComparer]::Ordinal)
    $active = $false
    $result = @{ version = ''; install_root = '' }
    $stream = [IO.File]::OpenRead($Path)
    try {
        if (-not $stream.CanSeek) { throw 'Bootstrap configuration must be a regular file' }
        $buffer = [byte[]]::new(1048577); $count = 0
        while ($count -lt $buffer.Length) {
            $read = $stream.Read($buffer, $count, $buffer.Length - $count)
            if (-not $read) { break }; $count += $read
        }
        if ($count -gt 1048576) { throw 'Bootstrap configuration exceeds 1 MiB' }
        $encoding = [Text.UTF8Encoding]::new($false, $true)
        $text = $encoding.GetString($buffer, 0, $count)
        if ($text.Length -gt 0 -and [int]$text[0] -eq 0xFEFF) { $text = $text.Substring(1) }
        $lines = $text -split '\r?\n'
    } finally { $stream.Dispose() }
    foreach ($raw in $lines) {
        if ($raw -match '^ *\t') { throw 'Tab indentation is unsupported' }
        $line = ''; $quote = ''; $previous = ''
        foreach ($character in $raw.ToCharArray()) {
            $c = [string]$character
            if ($quote) {
                if ($c -eq '\' -and $quote -eq '"') { throw 'Escaped double-quoted scalars are unsupported' }
                if ($c -eq $quote) { $quote = '' }
                $line += $c
            } else {
                if ($c -eq '#' -and ($line.Length -eq 0 -or $previous -match '[ \t]')) { break }
                if (($c -eq '"' -or $c -eq "'") -and ($line.Length -eq 0 -or $previous -match '[ \t:\[,]')) { $quote = $c }
                if ($c -match '[&*!|>]') { throw 'Anchors, aliases, tags and block scalars are unsupported' }
                $line += $c
            }
            $previous = $c
        }
        if ($quote) { throw 'Unclosed quoted scalar' }
        if ($line -match '^[ \t]*$') { continue }
        if ($line -match '^(---|\.\.\.|%)') { throw 'YAML document markers/directives are unsupported' }
        if ($line -match '^[^ ]') {
            if ($line -notmatch '^([A-Za-z_][A-Za-z0-9_-]*):(?:[ \t]+(.*))?$') { throw 'Use plain top-level mapping keys' }
            $key = $Matches[1]; $rest = [string]$Matches[2]
            if ($roots.ContainsKey($key)) { throw 'Duplicate top-level mapping key' }
            $roots[$key] = $true; $active = $key -ceq 'agenticlens'
            if ($active -and $rest.Trim([char[]]" `t")) { throw 'agenticlens must be a block mapping' }
            continue
        }
        if ($line -match '^ *<<:') { throw 'YAML merges are unsupported' }
        if (-not $active) { continue }
        if ($line -notmatch '^  ([A-Za-z_][A-Za-z0-9_-]*):(?:[ \t]+(.*))?$') { throw 'agenticlens children require exactly two spaces and plain keys' }
        $key = $Matches[1]; $value = ([string]$Matches[2]).Trim([char[]]" `t")
        if ($children.ContainsKey($key)) { throw 'Duplicate agenticlens key' }
        $children[$key] = $true
        if ($key -cne 'version' -and $key -cne 'install_root' -and $key -cne 'binary_dir') { throw 'Unknown agenticlens key' }
        if ($value.StartsWith('"') -or $value.StartsWith("'")) {
            $q = $value.Substring(0, 1)
            if ($value.Length -lt 2 -or -not $value.EndsWith($q)) { throw 'Unsupported quoted scalar' }
            $value = $value.Substring(1, $value.Length - 2)
            if ($value.Contains($q)) { throw 'Embedded quotes are unsupported' }
        } elseif ($value -match '[ \t#"'']') { throw 'Quote values containing spaces or comments' }
        if (-not $value -or $value -match '[\r\n]') { throw 'Empty or multiline scalar' }
        if ($key -ceq 'version') {
            if ($Local) { throw 'Local version overrides are prohibited; edit shared YAML' }
            if ($value.Length -gt 128 -or $value -cnotmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$') { throw 'Version must be an exact vMAJOR.MINOR.PATCH release tag' }
        } elseif ($key -ceq 'binary_dir') {
            if ($Local) { throw 'Local development selection overrides prohibited' }
            if ([Text.Encoding]::UTF8.GetByteCount($value) -gt 4096) { throw 'binary_dir exceeds path limit' }
        } elseif ([Text.Encoding]::UTF8.GetByteCount($value) -gt 4096 -or $value -notmatch '^(\/|~\/|[A-Za-z]:[\\\/]|\\\\[^\\]+\\[^\\]+)' -or $value -match '[<>|?*]') { throw 'install_root must be an absolute path or ~/path' }
        $result[$key] = $value
    }
    if (-not $Local -and -not $result.version) { throw 'Missing agenticlens.version' }
    return $result
}
try {
    $arguments = [Collections.Generic.List[string]]::new()
    foreach ($item in $args) { $arguments.Add([string]$item) }
    $project = '.'; $resolve = $false; $offline = $false
    while ($arguments.Count -gt 0) {
        if ($arguments[0] -ceq '--al-offline') { $offline = $true; $arguments.RemoveAt(0); continue }
        if ($arguments[0] -ceq '--al-resolve' -or $arguments[0] -ceq '--al-resolve-cache') { $resolve = $true; $arguments.RemoveAt(0); continue }
        if ($arguments[0] -ceq '--al-project') {
            if ($arguments.Count -lt 2) { throw '--al-project requires a directory' }
            $project = $arguments[1]; $arguments.RemoveRange(0, 2); continue
        }
        break
    }
    $root = Resolve-PhysicalDirectory $project $true
    $config = ''
    while ($true) {
        foreach ($suffix in @('yaml', 'yml', 'json')) {
            $candidate = [IO.Path]::Combine($root, ".agenticlens-config.$suffix")
            if ([IO.Directory]::Exists($candidate)) { continue }
            if ($null -ne (Get-Item -LiteralPath $candidate -Force -ErrorAction SilentlyContinue)) { if (-not [IO.File]::Exists($candidate)) { throw 'Configuration must be a regular file' }; $config = $candidate; break }
        }
        if ($config -or ($null -ne (Get-Item -LiteralPath ([IO.Path]::Combine($root, '.git')) -Force -ErrorAction SilentlyContinue))) { break }
        $parent = [IO.Directory]::GetParent($root)
        if ($null -eq $parent) { break }
        $root = $parent.FullName
    }
    if (-not $config) { throw 'Shared project YAML missing; explicit setup must write a version pin' }
    if ($config.EndsWith('.json')) { throw 'JSON bootstrap is unsupported; migrate shared configuration to .agenticlens-config.yaml' }
    # Match native capture flag arity; values such as ticket titles are data.
    if ($arguments.Count -gt 1 -and $arguments[0] -ceq 'capture') {
        $repoValue = ''; $pending = 0
        for ($index = 2; $index -lt $arguments.Count; $index++) {
            $item = $arguments[$index]
            if ($pending) { if ($pending -eq 2) { $repoValue = $item }; $pending = 0; continue }
            if ($item -ceq '--repo' -or $item -ceq '-repo') { $pending = 2; continue }
            if ($item -cmatch '^--?repo=(.*)$') { $repoValue = $Matches[1]; continue }
            if ($item -cmatch '^--?dry-run(?:=.*)?$') { continue }
            if ($item -ceq '--' -or -not $item.StartsWith('-')) { break }
            if (-not $item.Contains('=')) { $pending = 1 }
        }
        if ($repoValue) {
            $requested = Resolve-PhysicalDirectory $repoValue $true
            while ($true) {
                $found = $false
                foreach ($suffix in @('yaml', 'yml', 'json')) { if ([IO.File]::Exists([IO.Path]::Combine($requested, ".agenticlens-config.$suffix"))) { $found = $true } }
                if ($found -or ($null -ne (Get-Item -LiteralPath ([IO.Path]::Combine($requested, '.git')) -Force -ErrorAction SilentlyContinue))) { break }
                $parent = [IO.Directory]::GetParent($requested)
                if ($null -eq $parent) { break }; $requested = $parent.FullName
            }
            if (-not [string]::Equals($requested, $root, [StringComparison]::OrdinalIgnoreCase)) { throw 'Native --repo differs from selected project; supply matching --al-project DIR before capture' }
        }
    }
    $settings = Read-Bootstrap $config $false
    $localFile = [IO.Path]::Combine($root, '.agenticlens-config.local.yaml')
    if ([IO.Directory]::Exists($localFile)) { throw 'Local configuration must be a regular file' }
    if ([IO.File]::Exists($localFile)) {
        $overlay = Read-Bootstrap $localFile $true
        if ($overlay.install_root) { $settings.install_root = $overlay.install_root }
    }
    if ($env:AGENTIC_LENS_VERSION -and $env:AGENTIC_LENS_VERSION -cne $settings.version) { throw 'AGENTIC_LENS_VERSION conflicts with authoritative shared pin' }
    $homeDirectory = [Environment]::GetFolderPath('UserProfile')
    $installRoot = $settings.install_root
    if (-not $installRoot) { $installRoot = [IO.Path]::Combine($homeDirectory, '.agenticlens') }
    if ($installRoot.StartsWith('~/')) { $installRoot = [IO.Path]::Combine($homeDirectory, $installRoot.Substring(2)) }
    if ($installRoot -notmatch '^([A-Za-z]:[\\/]|\\\\)') { throw 'install_root must be absolute on the executing Windows platform' }
    $installRoot = Resolve-PhysicalDirectory $installRoot
    # This file is the native Windows entrypoint; other hosts use the POSIX wrapper.
    if ([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT) { throw 'Use the POSIX wrapper on this executing OS' }
    $architecture = [Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
    if ($architecture -cne 'X64') { throw 'Windows arm64/x86 artifacts are unavailable' }
    $native = [IO.Path]::Combine($installRoot, 'bin', $settings.version, 'windows-amd64', 'agentic-lens-windows-amd64.exe')
    if ($resolve) {
        $output = [Text.Encoding]::UTF8.GetBytes($settings.version + "`n" + $root + "`n" + $installRoot + "`n" + $native + "`n")
        $stdout = [Console]::OpenStandardOutput(); $stdout.Write($output, 0, $output.Length); $stdout.Flush()
        exit 0
    }
    if (-not [IO.File]::Exists($native)) { throw "Missing pinned native $native for $root; run explicit setup" }
    $stream = [IO.File]::OpenRead($native)
    try { $magic1 = $stream.ReadByte(); $magic2 = $stream.ReadByte() } finally { $stream.Dispose() }
    if ($magic1 -ne 77 -or $magic2 -ne 90) { throw 'Cache artifact is not native; explicit repair required' }
    exit (Invoke-NativeTransport $native $arguments.ToArray() ($offline -and $arguments.Count -gt 1 -and $arguments[0] -ceq 'capture' -and $arguments[1] -ceq 'check'))
} catch {
    $diagnostic = [Text.Encoding]::UTF8.GetBytes($_.Exception.Message + "`n")
    $stderr = [Console]::OpenStandardError(); $stderr.Write($diagnostic, 0, $diagnostic.Length); $stderr.Flush()
    exit 2
}
