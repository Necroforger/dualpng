# Dualpng
<!-- TOC -->

- [Dualpng](#dualpng)
- [Cmd/webui](#cmdwebui)
- [Cmd/dualpng](#cmddualpng)
    - [Installation](#installation)
    - [Usage](#usage)
    - [Examples](#examples)
        - [Default Options](#default-options)
        - [Thread mask](#thread-mask)
        - [Checkerboard mask](#checkerboard-mask)
    - [Flags](#flags)

<!-- /TOC -->
# Cmd/webui
Merge images with a drag and drop interface.
Download a version from the [releases](https://github.com/Necroforger/dualpng/releases)

visit http://localhost in your browser to view the ui. it will launch on port 80 by default.


![img](https://i.imgur.com/6JDBhgs.gif)

# Cmd/dualpng
## Installation
`go get github.com/Necroforer/dualpng/cmd/dualpng`

Or download a version from the [releases](https://github.com/Necroforger/dualpng/releases)

## Usage
`dualpng [flags] img1 img2`

img1 or img2 can be either a local file or a web address like
[https://avatars1.githubusercontent.com/u/16108486?v=4&s=46](https://avatars1.githubusercontent.com/u/16108486?v=4&s=460)

## Examples
### Default Options
`dualpng -w 1024 img1.png img2.png`
### Thread mask
`dualpng -r1="230" -r2="230-255" -g 1300 -w 1500 -m=[[0,1,0,1,1],[1,0,1,1,1],[1,1,1,1,0],[1,1,1,0,1],[1,1,1,0,1],[1,1,0,1,0]] img1.png img2.png`
### Checkerboard mask
`dualpng -r1="230" -r2="230-255" -g 1300 -w 1500 -m=[[1,1],[1,0]] img1.png img2.png`
## Flags
If only a width, or only a height is provided the missing field will be calculated to preserve the aspect ratio of the images.

| Flag | Type   | Description                                                                                            |
|------|--------|--------------------------------------------------------------------------------------------------------|
| w    | Uint   | Width to resize both images to                                                                         |
| h    | Uint   | Height to resize both images to                                                                        |
| m    | String | Mask matrix to use for masking images. (ex) `[[1, 1],[1,0]]` will create a checkerboard pattern        |
| r1   | String | Colour range for the first image (default: "0-240")                                                    |
| r2   | String | Colour range for the second image (default: "240-255)                                                  |
| g    | Uint   | gAMA value (default: 2300). The gAMA value is multiplied by 100,000. So a gAMA of 0.023 would be 2,300 |
| o    | String | Path of the output image (default: "output.png")                                                       |