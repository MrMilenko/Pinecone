<p align="center">
  <img src="https://raw.githubusercontent.com/MrMilenko/PineCone/main/images/cleet.png "width="200" />
</p>
<p align="center">I'm Cleet. Cletus T. Pine.</p>

# PineCone
* A content sniffer for the Original Xbox.
# How-To
* Download the id_database.json
* Download the appropriate binary for your platform.
* Working Directory should contain the binary, the json, and a TDATA and UDATA folder you want to process.
* Run your binary from the commandline. e.g: ./pinecone
# About
* Our buddy Harcroft has been keeping a rolling list of missing content for nearly 20 years.
* The idea of this software is to cut out as much of the manual digging as possible, and expand on it as a tool to archive this data.
# Hows this work?
* Drop UDATA and TDATA into a dump folder.
* Analyze the dump for userdata and DLC's, User Created Content, Content Update Files.
# Planned Features
* Disect Disk images
* Import archived files
* Implement updating using Github for title information.
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
