To package the current folder:

```
minigit
```

Optionally, you can also pass a directory:
```
minigit --dir=my_folder
```

To inspect an existing package, you can use:
```
minigit --inspect
```

### Format of a package entry

Each package entry is composed of an index followed by file contents.

```
<file_path> <file_hash> <file_size>
<file_path> <file_hash> <file_size>
~<file_content>
~<file_content>
```
