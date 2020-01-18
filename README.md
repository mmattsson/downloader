# downloader
Small Go program to download a file and visualize the receive bandwidth in a graph in a browser.

Run by
> go run main.go

Optional parameters are
> -r Use random data to test the visualization.  
> -url Pre-populate the input string with the given URL. The url can either start with 'ftp://', 'http://', or 'https://'

Once the program is running you can open an browser to localhost:8080 to see the visualization. Make sure a URL is entered in the input text field (either pre-populated or entered manually) and click 'Download'. 

Note that the data is downloaded but not saved. This program is intended to visualize download speed, not actually downloading data.

