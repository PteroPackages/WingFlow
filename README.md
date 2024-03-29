# WingFlow
WingFlow is a useful CLI tool for automatic uploading and deployment of projects to Pterodactyl servers.

## Installation
See the [releases page](https://github.com/PteroPackages/WingFlow/releases).

### From Source
```
git clone https://github.com/PteroPackages/WingFlow.git
go build -o build/ wflow.go
```

## Getting Started
1. Run `wflow init`
2. Add your credentials to the config file
3. Run `wflow run`
4. Congrats, you've uploaded and deployed your repository to your server!

## Commands
* `init`: Creates a new WingFlow configuration file in the current workspace. You can also specify a directory using the `--dir` flag.
* `check`: Runs valition checks on the local configuration file, or a specified one using the `--dir` flag.
* `run`: Fetches and deploys the configured project to Pterodactyl using Git.

## Contributing
1. [Fork this repo](https://github.com/PteroPackages/WingFlow/fork)!
2. Make a branch from `main` (`git branch -b <new feature>`)
3. Commit your changes (`git commit -am "..."`)
4. Open a PR here (`git push origin <new feature>`)

## Contributors
* [Devonte](https://github.com/devnote-dev) - creator and maintainer

This repository is managed under the MIT license.

© 2022-present PteroPackages
