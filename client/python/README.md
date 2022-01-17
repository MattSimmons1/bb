
# bb Python Client

### Install

First install the bb binary, then pip install:

    go install github.com/MattSimmons1/bb@latest
    pip install bb-python

If you don't have go installed, download the bb binary from the [releases page](https://github.com/MattSimmons1/bb/releases) 
and save to a location on your PATH or your working directory.

### Usage

Convert bb strings to Python lists/dicts (JSON objects):

```python
import bb

data = bb.convert("""
    
    ∆ = { type: pizza }
    
    3∆ 5∆ 5∆ 8∆ 2∆
    
""")

print(data)  # [{'type': 'pizza', 'quantity': 3}, {...
```

Convert files by providing a file path:

```python
import bb

data = bb.convert("path/to/file.bb.txt")
```

Convert in injection mode with `bb.extract`:

```python
import bb
documentation = bb.extract("my_code.py")
```
