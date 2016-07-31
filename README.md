goDash
==========

[![Codewake](https://www.codewake.com/badges/ask_question_flat_square.svg)](https://www.codewake.com/p/godash)

goDash is [Golang][1] based framework that lets you build beautiful dashboards.
This project is a "fork" of the original project [shopify/dashing][2] and [gigablah/dashing-go][3]

Key features:

* Works out of the box, no server, runtime or dependency requiered.
* Use premade widgets, or fully create your own with css, html, and js.
* Push Data from JIRA to your dashboard with a html attribute
* Auto-schedule and execute any script/binary file to feed data to your dashboard.
* Use the API to push data to your dashboards.

This dashing project was created at Shopify for displaying custom dashboards on TVs around the office.

# Getting Started
1. Get the app here https://github.com/vjeantet/goDash/releases
2. Start goDash   ```$ ./goDash```
3. Go to http://127.0.0.1:8080

	Your should see something like : 
	![alt tag](https://raw.githubusercontent.com/vjeantet/goDash/master/screenshot.png)	


On the first start goDash generates a demo dashboard and jobs to feed it.

# Settings
* default http port is 8080
	* set a environnement variable ```PORT``` to change this.
* default working directory is the path where it starts
	* set ```WEBROOT```env var to change this.
* default api TOKEN is empty
	* set ```TOKEN```env var to change this.


# Create a new dashboard
create a name_here.gerb file in the ```dashboards``` folder

* every 20s, goDash will switch to each dashboard it founds in this folder.
* you can group your dashboard in a folder.
	* example : ```dashboards/subfolder/dashboard1.gerb```  will be available to http://127.0.0.1:8080/subfolder/dashboard1. 
	* doDash will auto switch dashboards it founds in the sub folder.

## Customize layout
* modify ```dashboards/layout.gerb```
	* if you add a layout.gerb in a dashboards/subfolder it will be used by goDash when displaying a subfolder's dashboard.


# Feed data to your dashboard

## jobs folder usage
When you place a file in ```jobs``` folder Then goDash will immediatly execute and schedule it according to this convention : ```NUMBEROFSECONDS_WIDGETID.ext```
* the filename has 2 parts :
	* NUMBEROFSECONDS,  interval in seconds for each execution of this file.
	* WIDGETID, the ID of the widget on your dashboard.

The output of the executed file should be a json representing the data to send to your widget, see examples in ```jobs``` folder.

2 cli arguments are provided to each executed file
1. The url of the current running goDash
2. the token of the current running goDash API
3. 
You can use this if you want to send data to multiple widgets. (see example)

### PHP files
When you add a php file to the ```job``` folder, goDash will  assume ```php``` is available on your system, and will run it with it.

## HTTP call usage (dashing API)
```
curl -d '{ "auth_token": "YOUR_AUTH_TOKEN", "text": "Hey, Look what I can do!" } http://127.0.0.1:8080/widgets/YOUR_WIDGET_ID
```


## JIRA Jql and filters
Edit your .gerb dashboard to add jira attributes to your widget :

* ```jira-count-filter='17531'``` - goDash will search jiras with this filter and feed the widget with issues count.
* ```jira-count-jql='resolution is EMPTY'``` - goDash will search jiras with this JQL and feed the widget with issues count.
* ```jira-warning-over='10'``` - widget status will pass to warning when there is more dans 10 issues
* ```jira-danger-over='20'``` - widget status will pass to danger when there is more dans 20 issues

You don't need to restart goDash when editing gerb files to take changes into account.

### jira configuration
create a ```conf/jiraissuecount.ini``` file in goDash working directory.
* set url, username, password, interval in the file, 

```
url = "https://jira.atlassian.com/"
username = "" #if empty jira will be used anonymously
password =  ""
interval = 30
```


# Use your custom assets, widgets...
* goDash looks for assets in a ```public``` folder, when it can not found a file in this folder, it will use its embeded one.

## Widgets
To add a custom widget "Test"
* create a ```widgets``` folder in working directory
	* create a ```Test``` folder
		* add the ```Test.js```, ```Test.html```, ```Test.css``` files in it.

goDash will use them as soon as you set a widget with a ```data-view="Test"```

Be sure to look at the [list of third party widgets][4].


### COFFEESCRIPT ? SCSS ? JS ? CSS ?
* convert coffeescript to js : http://js2.coffee
* convert scss to css : http://www.sassmeister.com


[1]: http://golang.org
[2]: http://shopify.github.io/dashing
[3]: https://github.com/gigablah/dashing-go
[4]: https://github.com/Shopify/dashing/wiki/Additional-Widgets
