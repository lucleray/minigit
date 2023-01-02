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
<file_path> <file_hash> <version> <offset> <size>
<file_path> <file_hash> <version> <offset> <size>

<file_content><file_content>
```

For example:

```
file1	54caaafe12a7afbee7d0ac99cbf915fd	~	0	13
file2	64780afe7d95553571f4e2d387595244	~	13	13
sub/file1	312b7049cce1660d75b0728a7de2c0ea	495c690d445fe3a576283be7051f2904	0	31

file1_contentfile2_content
```