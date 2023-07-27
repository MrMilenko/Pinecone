/*Pinecone CPP Rewrite*/
#include <iostream>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <map>
#include <algorithm>
#include <cstring>
#include <cstdio>
#include <cstdlib>
#include <ctime>
#include <cmath>
#include <iomanip>
#include <openssl/sha.h>
#include <filesystem>
#include <rapidjson/document.h>
#include <rapidjson/istreamwrapper.h>

namespace fs = std::filesystem;

struct TitleUpdate
{
    std::string Name;
    std::string SHA1;
};

using ArchivedContent = std::map<std::string, std::string>;

struct TitleData
{
    std::string TitleName;
    std::vector<std::string> ContentIDs;
    std::vector<TitleUpdate> TitleUpdates;
    std::vector<TitleUpdate> TitleUpdatesKnown;
    std::vector<ArchivedContent> Archived;
};

struct Titles
{
    std::map<std::string, TitleData> Titles;
};

void loadJSONData(Titles &titles)
{
    std::string localPath = "data/id_database.json";
    std::cout << "Loading JSON data from " << localPath << std::endl;
    std::ifstream file(localPath);
    if (!file.is_open())
    {
        throw std::runtime_error("Failed to open file: " + localPath);
    }
    std::stringstream buffer;
    buffer << file.rdbuf();
    file.close();
    std::string json = buffer.str();

    rapidjson::Document jsonData;
    jsonData.Parse(json.c_str());

    if (jsonData.IsNull() || !jsonData.IsObject())
    {
        throw std::runtime_error("Failed to parse JSON or JSON is not an object");
    }

    const auto &jsonTitles = jsonData["Titles"];

    if (!jsonTitles.IsObject())
    {
        throw std::runtime_error("JSON does not contain 'Titles' object");
    }

    for (auto itr = jsonTitles.MemberBegin(); itr != jsonTitles.MemberEnd(); ++itr)
    {
        TitleData data;
        const auto &title = itr->value;

        std::string titleID = itr->name.GetString();
        std::transform(titleID.begin(), titleID.end(), titleID.begin(), ::tolower);

        if (title.HasMember("Title Name") && title["Title Name"].IsString())
        {
            data.TitleName = title["Title Name"].GetString();
        }

        if (title.HasMember("Content IDs") && title["Content IDs"].IsArray())
        {
            for (const auto &contentID : title["Content IDs"].GetArray())
            {
                data.ContentIDs.push_back(contentID.GetString());
            }
        }

        if (title.HasMember("Title Updates") && title["Title Updates"].IsArray())
        {
            for (const auto &update : title["Title Updates"].GetArray())
            {
                TitleUpdate titleUpdate;
                titleUpdate.Name = update["Name"].GetString();
                titleUpdate.SHA1 = update["SHA1"].GetString();
                data.TitleUpdates.push_back(titleUpdate);
            }
        }

        if (title.HasMember("Title Updates Known") && title["Title Updates Known"].IsArray())
        {
            for (const auto &update : title["Title Updates Known"].GetArray())
            {
                TitleUpdate titleUpdate;
                titleUpdate.Name = update["Name"].GetString();
                titleUpdate.SHA1 = update["SHA1"].GetString();
                data.TitleUpdatesKnown.push_back(titleUpdate);
            }
        }

        if (title.HasMember("Archived") && title["Archived"].IsArray())
        {
            for (const auto &archived : title["Archived"].GetArray())
            {
                ArchivedContent archivedContent;
                for (auto itr = archived.MemberBegin(); itr != archived.MemberEnd(); ++itr)
                {
                    archivedContent[itr->name.GetString()] = itr->value.GetString();
                }
                data.Archived.push_back(archivedContent);
            }
        }

        titles.Titles[titleID] = data;
    }
}


std::string getSHA1Hash(const std::string &filePath)
{
    std::ifstream file(filePath, std::ios::binary);
    if (!file.is_open())
    {
        throw std::runtime_error("Failed to open file: " + filePath);
    }
    std::vector<char> buffer(4096);
    SHA_CTX sha1;
    SHA1_Init(&sha1);
    while (file.read(buffer.data(), buffer.size()))
    {
        SHA1_Update(&sha1, buffer.data(), buffer.size());
    }
    SHA1_Update(&sha1, buffer.data(), file.gcount());
    std::vector<unsigned char> hash(SHA_DIGEST_LENGTH);
    SHA1_Final(hash.data(), &sha1);
    std::stringstream ss;
    for (const auto &byte : hash)
    {
        ss << std::hex << std::setw(2) << std::setfill('0') << static_cast<int>(byte);
    }
    return ss.str();
}

bool contains(const std::vector<std::string> &vec, const std::string &val)
{
    return std::find(vec.begin(), vec.end(), val) != vec.end();
}
void checkForContent(const std::string &directory, Titles &titles)
{
    // Check if directory exists
    if (!fs::exists(directory))
    {
        std::cerr << directory << " directory not found" << std::endl;
        return;
    }

    for (const auto &entry : fs::directory_iterator(directory))
    {
        if (entry.is_directory() && entry.path().filename().string().size() == 8)
        {
            std::string titleID = entry.path().filename().string();
            std::transform(titleID.begin(), titleID.end(), titleID.begin(), ::tolower);
            auto it = titles.Titles.find(titleID);

            if (it != titles.Titles.end())
            {
                std::cout << "Found folder for \"" << it->second.TitleName << "\".\n";

                std::string subDirDLC = entry.path().string() + "/$c";
                if (fs::exists(subDirDLC) && fs::is_directory(subDirDLC))
                {
                    // Handle $c subdirectories
                    for (const auto &subContent : fs::directory_iterator(subDirDLC))
                    {
                        if (subContent.is_directory())
                        {
                            std::string contentID = subContent.path().filename().string();
                            std::transform(contentID.begin(), contentID.end(), contentID.begin(), ::tolower);

                            std::cout << "Checking contentID: " << contentID << "\n";  // DEBUG

                            if (contains(it->second.ContentIDs, contentID))
                            {
                                // DEBUG
                                std::cout << "ContentID found in known IDs: " << contentID << "\n";
                                
                                // Add logic to check if content is archived or not
                                std::string archivedName;
                                for (const auto &archived : it->second.Archived)
                                {
                                    auto archived_it = archived.find(contentID);
                                    if (archived_it != archived.end())
                                    {
                                        archivedName = archived_it->second;
                                        break;
                                    }
                                }

                                // Add logic to list .xbe and .xbx files if unarchived content found
                                if (!archivedName.empty())
                                {
                                    std::cout << it->second.TitleName << " content found at: " << subContent.path().string() << " is archived (" << archivedName << ").\n";
                                }
                                else
                                {
                                    std::cout << it->second.TitleName << " has unarchived content found at: " << subContent.path().string() << "\n";
                                    for (const auto &file : fs::directory_iterator(subContent.path()))
                                    {
                                        if (file.is_regular_file())
                                        {
                                            std::string extension = file.path().extension().string();
                                            if (extension == ".xbe" || extension == ".xbx")
                                            {
                                                std::cout << "Found content.. " << file.path().filename().string() << "\n";
                                            }
                                            else
                                            {
                                                std::cout << "Found unknown file format: " << file.path().filename().string() << "\n";
                                            }
                                        }
                                    }
                                }
                            }
                            else
                            {
                                std::cout << it->second.TitleName << " unknown content found at: " << subContent.path().string() << "\n";
                            }
                        }
                    }
                }
                else
                {
                    std::cout << "No DLC Found for " << titleID << "..\n";
                }

                std::string subDirUpdates = entry.path().string() + "/$u";
                if (fs::exists(subDirUpdates) && fs::is_directory(subDirUpdates))
                {
                    // Handle $u subdirectories
                    for (const auto &file : fs::directory_iterator(subDirUpdates))
                    {
                        if (file.is_regular_file())
                        {
                            std::string extension = file.path().extension().string();
                            if (extension == ".xbe" || extension == ".xbx")
                            {
                                std::string filePath = file.path().string();
                                std::string fileHash;
                                try
                                {
                                    fileHash = getSHA1Hash(filePath);
                                }
                                catch (const std::exception &e)
                                {
                                    std::cerr << "Error computing SHA1 hash for file " << filePath << ": " << e.what() << std::endl;
                                    continue;
                                }

                                bool hashMatchFound = false;
                                for (const auto &knownUpdate : it->second.TitleUpdatesKnown)
                                {
                                    if (fileHash == knownUpdate.SHA1)
                                    {
                                        std::string name = knownUpdate.Name;
                                        size_t splitPosition = name.find(':');
                                        if (splitPosition != std::string::npos)
                                        {
                                            name = name.substr(splitPosition + 1);
                                        }
                                        std::cout << "Title update found for " << it->second.TitleName << " (" << titleID << ") (" << name << ")\n";
                                        std::cout << "Path: " << filePath << "\n";
                                        std::cout << "SHA1: " << fileHash << "\n";
                                        std::cout << "====================================================================================================\n";
                                        hashMatchFound = true;
                                        break;
                                    }
                                }

                                if (!hashMatchFound)
                                {
                                    std::cout << "No SHA1 hash matches found for file " << file.path().filename().string() << "\n";
                                    std::cout << "SHA1 for unknown content: " << fileHash << "\n";
                                    std::cout << "Path: " << filePath << "\n";
                                    std::cout << "====================================================================================================\n";
                                }
                            }
                        }
                    }
                }
                else
                {
                    std::cout << "No Title Updates Found in $u for " << titleID << "..\n";
                    std::cout << "====================================================================================================\n";
                }
            }
            else
            {
                std::cout << "Title ID " << titleID << " not present in JSON file.\n";
                std::cout << "We found a folder with the correct format, but it's not in the JSON file.\n";
                std::cout << "Please report this to the developer.\n";
                std::cout << "Path: " << directory << "\n";
                std::cout << "====================================================================================================\n";
            }
        }
    }
}

void printTitleStats(const TitleData &data)
{
    std::cout << "Title: " << data.TitleName << std::endl;
    std::cout << "Total number of Content IDs: " << data.ContentIDs.size() << std::endl;
    std::cout << "Total number of Title Updates: " << data.TitleUpdates.size() << std::endl;
    std::cout << "Total number of Known Title Updates: " << data.TitleUpdatesKnown.size() << std::endl;
    std::cout << "Total number of Archived items: " << data.Archived.size() << std::endl;
    std::cout << std::endl;
}

void printStats(const std::string &titleID, bool batch, const Titles &titles)
{
    if (batch)
    {
        for (const auto &[id, data] : titles.Titles)
        {
            std::cout << "Statistics for title ID " << id << ":" << std::endl;
            printTitleStats(data);
        }
    }
    else
    {
        auto it = titles.Titles.find(titleID);
        if (it == titles.Titles.end())
        {
            std::cout << "No data found for title ID " << titleID << std::endl;
            return;
        }
        std::cout << "Statistics for title ID " << titleID << ":" << std::endl;
        printTitleStats(it->second);
    }
}

int main(int argc, char *argv[])
{
    bool summarizeFlag = false;
    std::string titleIDFlag;
    bool fatxplorer = false;
    for (int i = 1; i < argc; i++)
    {
        if (std::strcmp(argv[i], "-summarize") == 0)
        {
            summarizeFlag = true;
        }
        else if (std::strncmp(argv[i], "-titleid=", 9) == 0)
        {
            titleIDFlag = argv[i] + 9;
        }
        else if (std::strcmp(argv[i], "-fatxplorer") == 0)
        {
            fatxplorer = true;
        }
        else if (std::strcmp(argv[i], "-help") == 0)
        {
            std::cout << "Usage of Pinecone:" << std::endl;
            std::cout << "  -summarize: Print summary statistics for all titles. If not set, checks for content in the TDATA folder." << std::endl;
            std::cout << "  -titleid: Filter statistics by Title ID (-titleID=ABCD1234). If not set, statistics are computed for all titles." << std::endl;
            std::cout << "  -fatxplorer: Use FATXPlorer's X drive as the root directory. If not set, runs as normal. (Windows Only)" << std::endl;
            std::cout << "  -help: Display this help information." << std::endl;
            return 0;
        }
    }

    Titles titles;
    try
    {
        loadJSONData(titles);
    }
    catch (const std::exception &e)
    {
        std::cerr << "Error loading JSON data: " << e.what() << std::endl;
        return 1;
    }

    std::cout << "Pinecone v0.3.1b" << std::endl;
    std::cout << "Please share output of this program with the Pinecone team if you find anything interesting!" << std::endl;
    std::cout << "====================================================================================================" << std::endl;

    if (!titleIDFlag.empty())
    {
        // if the titleID flag is set, print stats for that title
        printStats(titleIDFlag, false, titles);
    }
    else if (summarizeFlag)
    {
        // if the summarize flag is set, print stats for all titles
        printStats("", true, titles);
    }
    else if (fatxplorer)
    {
#ifdef _WIN32
        if (fs::exists("X:/"))
        {
            std::cout << "Checking for Content..." << std::endl;
            std::cout << "====================================================================================================" << std::endl;
            checkForContent("X:/TDATA", titles);
        }
        else
        {
            std::cout << "FatXplorer's X: drive not found" << std::endl;
        }
#else
        std::cout << "FatXplorer mode is only available on Windows." << std::endl;
#endif
    }
    else
    {
        // If no flag is set, proceed normally
        // Check if TDATA folder exists
        if (!fs::exists("dump/TDATA"))
        {
            std::cout << "TDATA folder not found. Please place TDATA folder in the dump folder." << std::endl;
            return 1;
        }
        std::cout << "Checking for Content..." << std::endl;
        std::cout << "====================================================================================================" << std::endl;
        checkForContent("dump/TDATA", titles);
    }

    return 0;
}
