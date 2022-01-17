
from setuptools import setup
import os

here = os.path.abspath(os.path.dirname(__file__))

# The text of the README file
README = open(here + "/README.md").read()

setup(
    name="bb-python",
    version="0.1.1",
    description="bb Python Client",
    long_description=README,
    long_description_content_type="text/markdown",
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
