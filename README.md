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
```
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
# Experimental Flags
* -fatxplorer: This flag will use a mounted E drive on partition X to scan.
* -update: This flag updates only the JSON. Useful between builds without major changes.
* -statistics: This will output statistics of the JSON, i.e totals.
* -titleid=ABCD1234: This will output the JSON details on a specific TitleID when provided.
# Example output
```
Local JSON file exists.
Loading JSON data...
Traversing directory structure...
Found folder for "Advent Rising".
Advent Rising has unarchived content found at: TDATA/4d4a0009/$c/4d4a000900000003
Title ID 50430001 not present in JSON file. May want to investigate!
Traversing directory structure for Title Updates...
TDATA/4d4a0009/$u/test.xbe: 87088e689b192c389693b3db38d5f26f2c4d55ae
```
