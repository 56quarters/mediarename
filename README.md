# mediarename

Rename media files based on their metadata.

## Build

TBD

## Usage

TBD

## Limitations

Files can only be [renamed](https://man7.org/linux/man-pages/man2/rename.2.html). The implication
of this is that the original files are renamed so if you want to preserve them, you must make a copy
before you rename them. Another implication is that files may only be renamed to a location on the
same filesystem partition as the original files.

For example, the following rename will work:

```
./mediarename tv --commit tt1234 ~/some/path ~/some/other-path
```

While this rename will fail:

```
./mediarename tv --commit tt1234 ~/some/path /mnt/media/some-other-device
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

