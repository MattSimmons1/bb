
from setuptools import setup
import pathlib

here = pathlib.Path(__file__).parent

# The text of the README file
README = (here / "README.md").read_text()

setup(
    name="bb",
    version="0.1.0",
    description="bb Python Client",
    long_description=README,
    packages=["bb"],
    author='Matt Simmons',
    author_email='',
    license="MIT",
    classifiers=[
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.9",
    ],
    url="https://mattsimmons1.github.io/bb",
    install_requires=[],
    package_data={"bb": ["bb"]},
    include_package_data=True,
)
