Face Verification With EverAI
=============================

Uses the [EverAI](https://ever.ai/) API to compare a face captured from the webcam to a reference image to determine 
whether the faces match.

Requires [GoCV](https://gocv.io/getting-started/linux/)

`LD_LIBRARY_PATH` might require an update after following the installation process above

```shell
export LD_LIBRARY_PATH=/usr/local/lib64:$LD_LIBRARY_PATH
```


Usage
-----

```shell
go run main.go -ref=/path/to/reference/image.jpg -host=everai.host
```


