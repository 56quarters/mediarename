# mediarename

Rename media files based on their metadata.

`mediarename` takes an IMDB ID for a TV show and path as input and renames the files
based on their metadata, from the [TVmaze](https://www.tvmaze.com/) API.

## Usage

```
./mediarename tv [<flags>] <id> <src> <dest>
```

To preview how files would be renamed for the TV show with IMDB ID `tt1234`:

```
./mediarename tv tt1234 ~/some-files ~/renamed-files
```

This will show you how each of the recognized TV show files in `~/some-files` will be
renamed into the destination directory `~/renamed-files`. This will _not actually rename
the files_.

In order to rename the files, you must provide the `--commit` flag to the command.

```
./mediarename tv --commit tt1234 ~/some-files ~/renamed-files
```

This will show you how each of the recognized TV show files in `~/some-files` will be
renamed into the destination directory `~/renamed-files` and then they _will be renamed_.

`mediarename` relies on season and episode numbers being in an expected format for each
file. It is required that each file includes these in the format (for example) `s01e03`
which indicates  that this file is season 1, episode 3. If a file does not include season
and episode number it will be skipped (not renamed) and a warning will be printed.

## Build

`mediarename` must be built from source using [Go](https://go.dev/). Once you have
installed Go, obtain the source code for `mediarename`:

```
git clone https://github.com/56quarters/mediarename.git
```

Then build from the root of the source code repository:

```
cd mediarename && make
```

This will create a `mediarename` binary in the repository that you can run:

```
./mediarename --help
```

## Limitations

Files can only be [renamed](https://man7.org/linux/man-pages/man2/rename.2.html). The implication
of this is that the original files are renamed so if you want to preserve them, you must make a copy
before you rename them. Another implication is that files may only be renamed to a location on the
same filesystem partition as the original files.

For example, the following rename will work:

```
./mediarename tv --commit tt1234 ~/some-files ~/renamed-files
```

While this rename will fail:

```
./mediarename tv --commit tt1234 ~/some-files /mnt/media/some-other-device
```

## Metadata API

TV Show metadata is fetched using the [TVmaze](https://www.tvmaze.com/) API. This API is free to use,
subject to some rate limits, and doesn't require an API key. Please do _not_ abuse the API and ruin things
for everyone. Information from the API is available under the [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/)
license. See the [API documentation](https://www.tvmaze.com/api) for more information.

## License

mediarename is available under the terms of the [GPL, version 3](LICENSE).

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in the work by you shall be licensed as above, without any
additional terms or conditions.

