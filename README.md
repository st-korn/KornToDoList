Experimental web-app ToDo-list by Stanislav Kornienko's conception.
You can try it in the cloud http://todo.works

![logo](https://github.com/st-korn/KornToDoList/raw/newplatform/static/favicon.png =64x64)

# Project Files
* `main.go` - backend server application, that implements interfaces for HTML-generation, HTTP static-files transfer, JSON-api for javascript and mobile applications.
* `templates/` - here are HTML-templates and JSON language packages. A file from this folder is used by the main server application, but not transmitted to clients directly via HTTP requests.
* `static/` - here are the static files that clients should receive via http requests: javascripts, icons, etc.
* `.graphics/` - in this folder the original images and the results of their processing are saved. All images received from third-party sites and services must necessarily contain a license, which allows their public commercial use.

# Environment variables
To run .go server-application you need to set these environment variables:

    SET MONGODB_ADDON_URI=mongodb://username:password@domain.com:port/databasename
    SET MONGODB_ADDON_DB=databasename
    SET PORT=port on which the web server listens

# How we name variables in Go?

We use both: lowerCamelCase or UpperCamelCase:
* If the variable is global and should be visible in more than one procedure, then use **UpperCamelCase**.
* For local variables that are used on short segments of the code, we use **lowerCamelCase**.
* For **type identifiers**, we use lowerCamelCase, which begins with the word "type": eg `typeTask`.
* All **fields of structures** are named in UpperCamelCase.
* We use **short names** of a small number of letters (eg `i`, `err`) if they are used in no more than several neighboring lines of code.

# How we name classes and id's in HTML DOM?

