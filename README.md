# GoScan
GoScan is a barebones proof-of-concept implementation of an document imaging server with the primary functionality of data extraction. 

## Prerequisites
* go 1.26.1
* The project uses [Tesseract OCR](https://github.com/tesseract-ocr/tesseract?tab=readme-ov-file#installing-tesseract) for obtaining data from images. A current limitation is that tesseract must be accessible through terminal. ```tesseract -v``` to verify your installation.
* Though not _required_, it is highly recommended to utilize this in conjunction with [GoScan-UI](https://www.github.com/bensoncarlb/GoScan-ui). This provides a (painful to look at) user interface for interacting with the server files and some limited server management.

## Usage
 To start up a new server: ```go run github.com/bensoncarlb/GoScan@latest```. This will also create the necessary directory structure in the current working directory. Once it is ready, it will be listening on ```localhost:8090``` for new requests.

 > Without any Document Types configured, any documents dropped into ```pickup``` will be processed as a Default type with all read data thrown into a single ```data``` field.

 Drop a png image into the ```pickup``` directory in the current directory. After 5 seconds it will begin being processed.
 
 > The file will remain in the pickup directory even once it has begun processing.

Once processing is finished, the image and a text file of read data will be available in ```output```.

## Limitations
This project is a limited proof of concept to learn go, so there are many limitations currently. 

The primary of which are:
* Only PNG files are supported
* All settings are hardcoded 

## Motivation
I needed a project to do while getting around to learning go and I have some experience being on the receiving end of similar systems. 

## Planned Future Work
* Use flags and/or a config file to control server behavior.
* Handle processing other image types, and potentially other document types such as PDF.
* New pickup and output modules such as APIs.
* More/better test cases.
* Code cleanup and documentation.
* Dealing with event and error logging with more than just println.