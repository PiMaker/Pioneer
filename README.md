# Pioneer

![GitHub stars](https://img.shields.io/github/stars/PiMaker/Pioneer.svg?style=social&label=Star)

### Problem

Building your own IoT devices is fun and often helps you solve real world problems. Controlling them often involves writing small applications and scripts that execute certain functions. That's all great, but at some point you probably want to get away from using a command line interface to call all your amazing Python-/Bash-/Whatever-Scripts.

### Meet Pioneer

A simple, reactive and secure web interface to call command line functions from everywhere, simple and intuitive!

### Features

* Single configuration file
* Clean, modern user interface
* User accounts
* SSL encryption out of the box
* One-time and toggle commands supported
* Schedule your commands to run at certain times
* Low performance requirements, runs perfectly on a Raspberry Pi
* Batteries included, single binary deploy (plus config.json)!

### Installation

Download the version matching your device/system below, create a file called `config.json` in the same directory (tip: copy and paste the example config from this repository to get the basic structure) and execute the binary! For best results, set it up so that the binary is executed at startup.

Alternatively, if you have a Go compiler installed you can just call `go get -v -u github.com/PiMaker/Pioneer`

### Download

| System type                                              |
| -------------------------------------------------------- |
| [darwin (32 bit)](static/Pioneer-darwin-10.6-386?raw=true)        |
| [darwin (64 bit)](static/Pioneer-darwin-10.6-amd64?raw=true)      |
| [linux (32 bit)](static/Pioneer-linux-386?raw=true)               |
| [linux (64 bit)](static/Pioneer-linux-amd64?raw=true)             |
| [linux (ARM v5)](static/Pioneer-linux-arm-5?raw=true)             |
| [linux (ARM v6) (Raspberry Pi 1)](static/Pioneer-linux-arm-6?raw=true)             |
| [linux (ARM v7) (Raspberry Pi 2+)](static/Pioneer-linux-arm-7?raw=true)             |
| [linux (ARM64)](static/Pioneer-linux-arm64?raw=true)              |
| [linux (MISP64)](static/Pioneer-linux-mips64?raw=true)            |
| [linux (MIPS64le)](static/Pioneer-linux-mips64le?raw=true)        |
| [Windows (32 bit)](static/Pioneer-windows-4.0-386.exe?raw=true)   |
| [Windows (64 bit)](static/Pioneer-windows-4.0-amd64.exe?raw=true) |

Thanks [XGo](https://github.com/karalabe/xgo)!

### Configuration

Look at the file `config.json` in this repository's root folder for documentation and an example.

### Screenshots

![screenshot2](static/screenshot2.jpeg)
![screenshot1](static/screenshot1.jpeg)

Note: To get the amazing `htop` background image, you have to use the live background feature. Don't ask me how exactly I set it up though, this was one of those "it's almost midnight I want to do something fun" ideas, it quickly turned into a garbled mess though (as one would expect).

### TODO

* Documentation
* Testing

Note that this was started as a small side project, so the code is rather messy right now. It does work though, I've never had it crash on me after about half a year of continuous usage on a Raspberry Pi 2.

### License

This project is licensed under the MIT License. Look at [LICENSE](LICENSE) for further details.
