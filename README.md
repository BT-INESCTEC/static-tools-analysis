# Analysis of Static Analysis Tools

## Setup

1. Clone the repository with the submodule and navigate to the project directory.

```bash
git clone https://github.com/BT-INESCTEC/static-tools-analysis --recurse-submodules
cd static-tools-analysis
```

2. If you are using NixOS, there is a flake.nix available to setup the dev environment. You can use it with the following command:

```bash
nix develop
```

Otherwise, install the following dependencies manually:

- Python 3.11 or higher
- Poetry 
- Go
- Zizmor
- NodeJS
- [Ades](https://github.com/ericcornelissen/ades#installation)

3. Install project dependencies using Poetry.
```bash
poetry install
```

4. Install Argus dependencies using Poetry.

```bash
cd argus
poetry install
cd ..
```

5. Install CodeQL

```bash
# 1. Create the target directory
mkdir -p ~/codeql_home

# 2. Download the latest CodeQL CLI bundle (Linux x64)
cd /tmp
curl -L -o codeql-linux64.zip \
  https://github.com/github/codeql-cli-binaries/releases/latest/download/codeql-linux64.zip

# 3. Extract into ~/codeql_home/, produces ~/codeql_home/codeql/codeql
unzip codeql-linux64.zip -d ~/codeql_home

# 4. Verify
ls -l ~/codeql_home/codeql/codeql
~/codeql_home/codeql/codeql --version

# 5. Install CodeQl Packages
cd argus/qlqueries
~/codeql_home/codeql/codeql pack install
```

