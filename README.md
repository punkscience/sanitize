# Sanitize

A command-line tool that recursively walks a folder tree and renames directories to be compatible with Windows naming conventions.

## Features

- Processes folders from lowest level to highest level (bottom-up traversal)
- Removes invalid Windows characters: `< > : " | ? * \ /`
- Removes control characters (ASCII 0-31)
- Trims trailing spaces and periods
- Handles Windows reserved names (CON, PRN, AUX, NUL, COM1-COM9, LPT1-LPT9)
- Converts Unicode/non-ASCII characters to closest ASCII equivalents
- Enforces 255-character length limit
- Handles name collisions by appending numbers
- Dry-run mode to preview changes
- Verbose output for detailed progress

## Usage

```bash
# Sanitize current directory
sanitize.exe

# Sanitize specific directory
sanitize.exe -path "C:\path\to\directory"

# Dry run (preview changes without making them)
sanitize.exe -path "C:\path\to\directory" -dry-run

# Verbose output
sanitize.exe -path "C:\path\to\directory" -verbose

# Combine flags
sanitize.exe -path "C:\path\to\directory" -dry-run -verbose
```

## Options

- `-path string`: Root path to sanitize (default: current directory)
- `-dry-run`: Show what would be renamed without making changes
- `-verbose`: Enable verbose output

## Examples

### Before:
```
folder/
├── unicode_ñoñó/
├── ending_with_period./
├── CON/
├── bad<chars>/
└── very_long_name_that_exceeds_255_characters.../
```

### After:
```
folder/
├── unicode_nono/
├── ending_with_period/
├── CON_/
├── bad_chars/
└── very_long_name_that_exceeds_255_cha.../
```

## Windows Folder Naming Rules

The tool enforces these Windows compatibility rules:

1. Cannot contain: `< > : " | ? * \ /`
2. Cannot contain control characters (ASCII 0-31)
3. Cannot end with space or period
4. Cannot exceed 255 characters
5. Cannot use reserved names: CON, PRN, AUX, NUL, COM1-COM9, LPT1-LPT9
6. Cannot be empty or contain only spaces
7. Unicode characters are converted to ASCII equivalents

## Building

```bash
go build -o sanitize.exe
```

## Safety

- Always test with `-dry-run` first
- The tool processes folders from deepest to shallowest to avoid path conflicts
- Handles name collisions by appending numbers (_1, _2, etc.)
- Skips inaccessible directories with warnings