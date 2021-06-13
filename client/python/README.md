
# bb Python Client

### Install

    pip install client/python
    
or

    pip install "git+ssh://git@github.com/MattSimmons1/bb.git#subdirectory=client/python"
    
    
### Usage

Convert bb strings to Python lists/dicts (JSON objects)

```python
import bb

data = bb.convert("""
    
    ∆ = { type: pizza }
    
    3∆ 5∆ 5∆ 8∆ 2∆
    
""")

print(data)  # [{'type': 'pizza', 'quantity': 3}, {...
```

Convert file paths

```python
import bb

data = bb.convert("path/to/file.bb.txt")
```

Convert in injection mode with `bb.extract` 

```python
import bb
documentation = bb.extract("my_code.py")
```