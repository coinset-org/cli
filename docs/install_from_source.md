This document will describe how to build the Coinset CLI tool from source. If you are looking for a binary version of the tool, see the latest [release](https://github.com/coinset-org/cli/releases).

Regardless of which operating system you are using, you will need to use the `git` command line tool. If you do not have `git` installed, follow the [instructions](https://git-scm.com/downloads) to install it before continuing.

### MacOS

1. Verify that Go/golang is installed by running

```bash
go
```

You should see a usage statement. If instead you see an error, see the [Go website](https://go.dev/doc/install) for instructions to install Go.

2. Clone this repository

```bash
git clone https://github.com/coinset-org/cli.git
```

3. Change to the command line directory

```bash
cd cli/cmd/coinset
```

4. Build coinset

```bash
go build
```

5. To verify that the installation was successful, run a command, for example

```bash
./coinset get_blocks 123 125
```

This should return info from the requested blocks.

---

### Linux

1. On Debian/Ubuntu, install Go/golang by running

```bash
sudo apt install golang-go
```

To install Go on other distributions, see the [Go website](https://go.dev/doc/install) for instructions.

2. Clone this repository

```bash
git clone https://github.com/coinset-org/cli.git
```

3. Change to the command line directory

```bash
cd cli/cmd/coinset
```

4. Build coinset

```bash
go build
```

5. To verify that the installation was successful, run a command, for example

```bash
./coinset get_blocks 123 125
```

This should return info from the requested blocks.

---

## Windows

### Golang Installation

You will need to use Go/golang for installing coinset. To install golang, Chocolatey (similar to `apt-get` on Linux) is recommended. To install Chocolatey and golang, open a Powershell window [as an administrator](https://www.supportyourtech.com/tech/how-to-run-powershell-as-admin-windows-11-a-step-by-step-guide/), then run the following:

1. Set up an execution policy:

```powershell
Set-ExecutionPolicy -Scope CurrentUser
```

2. You will be prompted to enter a policy. Enter the following and press `enter`:

```powershell
RemoteSigned
```

Enter `y` to allow the changes to take effect.

3. To verify this change, list all policies:

```powershell
Get-ExecutionPolicy -List
```

Among the list, you should see `CurrentUser    RemoteSigned`.

4. Create a new web client object:

```powershell
$script = New-Object Net.WebClient
```

5. Install Chocolatey:

```powershell
iwr https://chocolatey.org/install.ps1 -UseBasicParsing | iex
```

You should see several lines of output, including `Installing Chocolatey on the local machine`.

6. Close your Powershell window and open a new one as administrator.

7. Upgrade Chocolatey:

```powershell
choco upgrade chocolatey
```

You should now be using the latest version.

8. Install golang:

```powershell
choco install -y golang
```

8. Close your Powershell window and open a new one (not as administrator).

9. Verify that golang is installed and accessible:

```powershell
go version
```

You should be shown the current version, for example `go version go1.23.1 windows/amd64`.

### Coinset Installation

1. Make and change to a directory for the Coinset repository

```powershell
mkdir coinset-org
```

```powershell
cd coinset-org
```

2. Clone this repository

```powershell
git clone https://github.com/coinset-org/cli.git
```

3. Change to the command line directory

```powershell
cd .\cli\cmd\coinset
```

4. Build coinset

```powershell
go build
```

5. To verify that the installation was successful, run a command, for example

```powershell
.\coinset get_blocks 123 125
```

Be sure to include the `.\` before the `coinset` command. This should return info from the requested blocks. If you see extra characters such as `←[37m` before each line, you are likely either running Powershell as an administrator (not recommended), or are not using the latest version (the latest version is recommended).

