# App

Name: sanitize

# Brief

A command-line app which recursively walks a folder tree from its lowest level to its highest level renaming folders to names which are compatible with Windows according to the Windows rules below.

# Windows folder rules

- Cannot contain any of these characters: `< > : " | ? * \`
- Cannot contain control characters (ASCII 0-31)
- Cannot end with a space or period
- Cannot be longer than 255 characters
- Cannot use these reserved names: CON, PRN, AUX, NUL, COM1-COM9, LPT1-LPT9 (case insensitive)
- Cannot be empty or contain only spaces
- Forward slash `/` is treated as a path separator, not allowed in names
- Backslash `\` is the path separator and not allowed in names
- All unicode or non-ASCII characters must be translated to their closest compatible ASCII characters.

