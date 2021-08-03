
import os


def which():
    """replicates the functionality of Unix which. Used to locate the go binary."""
    for path in os.getenv("PATH").split(os.path.pathsep):
        full_path = path + os.sep + "bb"
        if os.path.exists(full_path):
            return full_path
    return None
