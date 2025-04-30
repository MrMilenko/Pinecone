<p align="center">
  <img src="https://raw.githubusercontent.com/MrMilenko/PineCone/main/images/old-cleet.png "width="200" />
</p>
<p align="center">I'm Cleet. Cletus T. Pine.</p>

## Pinecone

**Project Status: Stable**

Pinecone is no longer under active development. The application itself is stable and functional. Only critical bug fixes or updates to the DLC/Title Update JSON may be made as needed.

For ongoing preservation efforts, we recommend using **XCAT**, which has become the community standard: [https://consolemods.org/wiki/Xbox:XCAT](https://consolemods.org/wiki/Xbox:XCAT)

Small patches may still be made to improve usability and fix bugs, especially for those of you with stacks of old drives still waiting to be scanned :)

## A Thank You to the Community

Pinecone began as an effort to rescue lost content from aging original Xbox hard drives. It grew into a utility that helped others do the same, preserving DLC, title updates, and (occasionally) other forgotten data before those drives were wiped, trashed, or sold.

This project was built on the work and research of many in the Xbox preservation scene but one name deserves special recognition:

**Harcroft** laid the groundwork. His years of documentation on content formats, Title Update structures, and obscure folder layouts made Pinecone possible.

## Why It’s “Done”

The preservation landscape has improved dramatically. With tools like [**XCAT**](https://consolemods.org/wiki/Xbox:XCAT), the process is now faster, more automated, and hardware-friendly.

You can now:
- Scan consoles directly
- Upload unarchived content for analysis
- Use modern drive unlocking via [**PrometheOS**](https://github.com/Team-Resurgent/PrometheOS-Firmware)

Special thanks to **Crunchbite**, **xbox7887**, **Siktah**, **SkyeHDD**, and **Team Resurgent** for these advancements.

Pinecone served its purpose and now Cleet can rest.


## If You're Just Finding This

The full source code and binaries will remain available here on GitHub.  
You're welcome to fork it, use it, or build on it however you see fit.

Thanks to everyone who contributed, tested, or shared Pinecone.  
You helped preserve a part of Xbox history that would have otherwise been lost.

## Community Links

**Discords**  
- [ConsoleMods Wiki Discord](https://discord.gg/x5vEnkR4C8)
- [Xbox-Scene Discord](https://discord.gg/xbox-scene)

**Sites**  
- [ConsoleMods.org](https://consolemods.org/wiki/Main_Page)
- [Digiex.net](https://digiex.net/downloads/download-center-2-0/xbox-content/)

**Wikis**   
- [XCAT Wiki](https://consolemods.org/wiki/Xbox:XCAT)  
- [PrometheOS Wiki](https://consolemods.org/wiki/Xbox:PrometheOS)

# PineCone

- A content discovery tool, for original Xbox DLC and Title Updates.

# How-To

- Download the id_database.json
- Download the appropriate binary for your platform.
- Working Directory should look like this:

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

- Run your binary from the commandline. e.g: ./pinecone (or pinecone.exe) (optional flags: -fatxplorer (Windows only, mount E as X in fatxplorer))

# About

- Our buddy Harcroft has been keeping a rolling list of missing content for nearly 20 years.
- The idea of this software is to cut out as much of the manual digging as possible, and expand on it as a tool to archive this data.

# Hows this work?

- Drop UDATA and TDATA into a dump folder.
- Analyze the dump for userdata and DLC's, User Created Content, Content Update Files.
- (Optional) Analyze the dump for Homebrew content in a C E F G folder structure.

# Todo

- Disect Disk images
- Import archived dumps
- Export output for easy viewing
- Add more flags for more specific searches
- Create "Homebrew" JSON file to identify homebrew content.
- Beautify output, to make it easier on the eyes.

# Flags

- `-f`/`--fatxplorer`: This flag will use a mounted E drive on partition X to scan.
- `-u`/`--update`: This flag updates only the JSON. Useful between builds without major changes.
- `-s`/`--statistics`: This will output statistics of the JSON, i.e totals.
- `-tID=ABCD1234`/`--titleid=ABCD1234`: This will output the JSON details on a specific TitleID when provided.
- `-l=path/to/dump`/`--location=path/to/dump`: Specify the directory where your dump is located
- `-g={true/false}`/`--gui={true/false}`: Enable the GUI interface (default = true)

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

## Dependencies

- `go` 1.21.5 or later
- `fyne` 2.4.3 or later
- `gofumpt` 0.6.0 or later (our preferred formatter)

## Build instructions

1. Install the `fyne` CLI tool

```sh
go install fyne.io/fyne/v2/cmd/fyne@latest
```

2. Install `gofumpt` CLI tool

```sh
go install mvdan.cc/gofumpt@latest
```

3. Run `go mod tidy` in the root directory to install all dependencies
4. Run `go build .`. WARNING: First compile will take a long time. Be patient!
