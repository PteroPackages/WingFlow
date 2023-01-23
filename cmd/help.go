package cmd

var initCmdHelp = `Creates a new config file in the current workspace. If one already exists, the
command will error unless the '--force' flag is included.`

var checkCmdHelp = `Runs validation checks on the config file, including HTTP checks for the panel
URL. To skip the HTTP checks, include the '--dry' flag.`

var runCmdHelp = `Fetches and deploys the configured project to the Pterodactyl server.

Process:
1. Checks on the configuration file are ran
2. Panel and repository connections are tested
3. Repository files are fetched and sorted according to the config
4. The files are written into a .tar.gz archive
5. Server files and signal options are executed for the server
6. The archive is uploaded and unpacked in the server`
