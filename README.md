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

