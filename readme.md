### Package a folder

```
minigit package <directory>
```

### Inspect a packaged folder

```
minigit inspect <directory>
```

### Format of package entry

Each package entry is composed of an index followed by file contents.

```
<file_path> <file_hash> <file_size>
<file_path> <file_hash> <file_size>
~<file_content>
~<file_content>
```
