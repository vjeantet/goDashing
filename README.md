goDash
==========

A working [Go][1] port of [shopify/dashing][2] without dependency.

# Features
* embeded web assets (javascript / css / ...) 
	* goDash will use your assets if they exists on disk. 
* auto schedule and run jobs you place in ```jobs/``` folder

# 1 minute start
Download a binary accordingly to your system : https://github.com/vjeantet/goDash/releases

## Start goDash :
```
$ ./goDash
```
* goDash will listen on port 8080
* goDash will create a example dashboard and jobs to feed it
	* if a ```dashboards``` folder already exists, it will not create it.
	* if a ```jobs``` folder already exists, it will not create it.

## Enjoy :
Open your browser, go to http://127.0.0.1:8080



![alt tag](https://raw.githubusercontent.com/vjeantet/goDash/master/screenshot.png)



# Settings
Change some settings with env variables
* ```PORT``` to choose which port to listen to
* ```WEBROOT``` to change the goDash working directory
* ```TOKEN``` to set a token to use with Dashing API


# Create a new dashboard
create a name_here.gerb file in the ```dashboards``` folder

* every 20s, goDash will switch to each dashboard it founds in this folder.
* you can group your dashboard in a folder.
	* example : ```dashboards/subfolder/dashboard1.gerb```  will be available to http://127.0.0.1:8080/subfolder/dashboard1. 
	* doDash will auto switch dashboards it founds in the sub folder.

## Customize layout
* modify ```dashboards/layout.gerb```
	* if you add a layout.gerb in a dashboards/subfolder it will be used by goDash when displaying a subfolder's dashboard.


# Feed data to your dashboard from jobs, http call or jira

## Feed data to your dashboard with jobs
When you place a file in ```jobs``` folder Then goDash will immediatly execute and schedule it according to this convention : ```NUMBEROFSECONDS_WIDGETID.ext```
* filename has 2 parts :
	* NUMBEROFSECONDS,  interval in seconds for each execution of this file.
	* WIDGETID, the ID of the widget on your dashboard.
* if the file is a php file, it will be run assuming ```php``` is available on your system.
	* others extentions with be executed directly.

The output of the executed file should be a json representing the data to send to your widget, see examples in ```jobs``` folder.

2 arguments are provided to each executed file
* The url of the current running goDash
* the token of the current running goDash API

You can use this if you want to send data to multiple widgets. (see example)

## Feed data to your dashboard with a http call
```
curl -d '{ "auth_token": "YOUR_AUTH_TOKEN", "text": "Hey, Look what I can do!" } http://127.0.0.1:8080/widgets/YOUR_WIDGET_ID
```


## Feed data to your dashboard from a JIRA instance
create a ```conf/jiraissuecount.ini``` file in working directory.
* set url, username, password, interval in the file, 

```
url = "https://jira.atlassian.com/"
username = "" #if empty jira will be used anonymously
password =  ""
interval = 30
```

add theses attributes to your Number widget in the dashboard .gerb file.

* search mode (use filter or jql)
	* ```jira-count-filter='17531'`` - goDash will search jiras with this filter and feed the widget with issues count.
	* ```jira-count-jql='resolution is EMPTY'`` - goDash will search jiras with this JQL and feed the widget with issues count.
* status (optional)
	* ```jira-warning-over='10'`` - widget status will pass to warning
	* ```jira-danger-over='20'`` - widget status will pass to danger


You don't need to restart goDash when editing gerb files to take changes into account.

# Use your custom assets, widgets...
* goDash looks for assets in a ```public``` folder, when it can not found a file in this folder, it will use its embeded one.

## Widgets
To add a custom widget "Test"
* create a ```widgets``` folder in working directory
	* create a ```Test``` folder
		* add the ```Test.js```, ```Test.html```, ```Test.css``` files in it.

goDash will use them as soon as you set a widget with a ```data-view="Test"```


### COFFEESCRIPT ? SCSS ? JS ? CSS ?
* convert coffeescript to js : http://js2.coffee
* convert scss to css : http://www.sassmeister.com




Credits
-------

goDash is a Fork from [github.com/gigablah/dashing-go][3]



[1]: http://golang.org
[2]: http://shopify.github.io/dashing
[3]: https://github.com/gigablah/dashing-go

