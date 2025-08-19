# githookd
A shell script that simplifies the management of Git hooks for a project. Instead of requiring developers to manually place executable scripts in the non-versioned `.git/hooks` directory, this tool allows hooks to be managed in a version-controlled directory (e.g., `.githooks/`) and automatically symlinks them into place.
