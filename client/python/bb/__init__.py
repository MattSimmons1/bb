
import subprocess
import json
import os
from typing import Any

from .util import which

here = os.path.abspath(os.path.dirname(__file__))  # something like site-packages/bb
workdir = os.path.abspath("")

#
# setup
#

HELP_MSG = "Install bb with: go get github.com/MattSimmons1/bb, " \
           "or download the binary and add it to your PATH or current working directory."

BB_PATH = which()  # look for existing bb installation

if BB_PATH is None:
    if os.path.isfile(f"{workdir}/bb"):  # look for bb in the working directory
        BB_PATH = f"{workdir}/bb"
    else:
        import platform
        # Determine the correct platform
        system = platform.system()
        arch, _ = platform.architecture()

        if system == 'Linux' and arch == '64bit':
            BB_PATH = f"{here}/lib/linux_386/bb"

        elif system == "Darwin" and arch == "64bit":
            BB_PATH = f"{here}/lib/darwin_amd64/bb"

        if BB_PATH is None:
            raise EnvironmentError(f"bb binary was not found for {system} {arch}! {HELP_MSG}")

if not os.access(BB_PATH, os.X_OK):
    os.chmod(BB_PATH, 0o777)
    assert os.access(BB_PATH, os.X_OK), "Cannot get permission to execute bb binary"

try:
    p = subprocess.Popen([BB_PATH], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
except OSError:
    raise EnvironmentError(f"The bb binary could not be executed. {HELP_MSG}")


#
#
#


class BBDecodeError(Exception):
    def __init__(self, message="Input is not valid bb."):
        self.message = message
        super().__init__(self.message)


def convert(input: str, definitions: str = None, injection_mode: bool = False) -> Any:
    """Convert bb syntax to a json object.

    :param input: bb string or file path.
    :param definitions: bb string or file path containing type definitions to use.
    :param injection_mode: If true, only bb found within comments will be parsed. Same as using bb.extract().
    :return: List of JSON objects representing the input.
    """
    cmd = [BB_PATH, input]

    if definitions is not None:
        cmd.append("-d")
        cmd.append(definitions)

    if injection_mode:
        cmd.append("-i")

    p = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    res = "\n".join(iter(p.stdout.readline, ""))
    if res == "":
        raise BBDecodeError
    res = json.loads(res)
    return res


def extract(input: str, definitions: str = None) -> Any:
    """Extract injected bb from within the comments of another language.

    :param input: bb string or file path.
    :param definitions: bb string or file path containing type definitions to use.
    :return: List of JSON objects representing the input.
    """
    return convert(input, definitions=definitions, injection_mode=True)


if __name__ == '__main__':  # unit tests

    assert convert("hello 1 2") == ['hello', 1, 2]
    assert convert("3.4∆", definitions="∆ = { cooleh: fooleh }") == {'cooleh': 'fooleh', 'quantity': 3.4}
