<p align="center">
  <img src="https://raw.githubusercontent.com/MrMilenko/PineCone/main/images/cleet.png "width="200" />
</p>
<p align="center">I'm Cleet. Cletus T. Pine.</p>

# PineCone

* A content discovery tool, for original Xbox DLC and Title Updates.

# How-To

* Download the id_database.json
* Download the appropriate binary for your platform.
* Working Directory should look like this:

```sh
PineCone
|-- pinecone binary
|
|-- data
| |-- id_database.json
|
|-- dump
| |-- TDATA
| |-- UDATA
| |-- C (optional)
| |-- E (optional)
| |-- F (optional)
| |-- G (optional)
```

* Run your binary from the commandline. e.g: ./pinecone (or pinecone.exe) (optional flags: -fatxplorer (Windows only, mount E as X in fatxplorer))

# About

* Our buddy Harcroft has been keeping a rolling list of missing content for nearly 20 years.
* The idea of this software is to cut out as much of the manual digging as possible, and expand on it as a tool to archive this data.

# Hows this work?

* Drop UDATA and TDATA into a dump folder.
* Analyze the dump for userdata and DLC's, User Created Content, Content Update Files.
* (Optional) Analyze the dump for Homebrew content in a C E F G folder structure.

# Todo

* Disect Disk images
* Import archived dumps
* Export output for easy viewing
* Add more flags for more specific searches
* Create "Homebrew" JSON file to identify homebrew content.
* Beautify output, to make it easier on the eyes.

# Flags

* `-f`/`--fatxplorer`: This flag will use a mounted E drive on partition X to scan.
* `-u`/`--update`: This flag updates only the JSON. Useful between builds without major changes.
* `-s`/`--statistics`: This will output statistics of the JSON, i.e totals.
* `-tID=ABCD1234`/`--titleid=ABCD1234`: This will output the JSON details on a specific TitleID when provided.
* `-l=path/to/dump`/`--location=path/to/dump`: Specify the directory where your dump is located
* `-g={true/false}`/`--gui={true/false}`: Enable the GUI interface (default = true)

# Example output

```sh
Pinecone v0.5.0
Please share output of this program with the Pinecone team if you find anything interesting!
Checking for Content...
====================================================================================================
============================================== Halo 2 ==============================================
    Content is known and archived Bonus Map Pack
    Content is known and archived Killtacular Pack
    Content is known and archived Maptacular Pack
    Content is known and archived Blastacular Pack
============================================ File Info =============================================
    Title update found for Halo 2 (4d530064) (0000000300000803:RF English Update 5)
    Path: dump\TDATA\4d530064\$u\default.xbe
    SHA1: f1cc1ae660161f4439fc29ee131310a86e326447

```

# Building from source

## Dependancies

* `go` 1.21.5 or later
* `fyne` 2.4.3 or later
* `gofumpt` 0.6.0 or later (our prefered formatter)

## Build instructions

1. Install the `fyne` CLI tool

```sh
go install fyne.io/fyne/v2/cmd/fyne@latest
```

2. Install `gofumpt` CLI tool

```sh
go install mvdan.cc/gofumpt@latest
```

3. Run `go mod tidy` in the root directory to install all depandancies
4. Run `go build .`. WARNING: First compile will take a long time. Be patient!

## Packaging for Release

Using our Makefile, run:

```sh
make build-{OS-NAME}
```

Where `OS-NAME` is either `win`, `linux`, or `mac` for your operating system of choice.

`FyneApp.toml` can be modified to change a variety of build variables, which you can find in the [fyne docs](https://docs.fyne.io/).
