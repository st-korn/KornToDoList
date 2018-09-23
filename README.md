Experimental web-app ToDo-list by Stanislav Kornienko's conception.
You can try it in the cloud http://todo.works

<img src="https://github.com/st-korn/KornToDoList/raw/newplatform/static/favicon.png" width="64">

# Project Files
* `main.go` - backend server application, that implements interfaces for HTML-generation, HTTP static-files transfer, JSON-api for javascript and mobile applications.
* `templates/` - here are HTML-templates and JSON language packages. A file from this folder is used by the main server application, but not transmitted to clients directly via HTTP requests.
* `static/` - here are the static files that clients should receive via http requests: javascripts, icons, etc.
* `.graphics/` - in this folder the original images and the results of their processing are saved. All images received from third-party sites and services must necessarily contain a license, which allows their public commercial use.

# Environment variables
To run .go server-application you need to set these environment variables:

    SET MONGODB_ADDON_URI=mongodb://username:password@domain.com:port/databasename
    SET MONGODB_ADDON_DB=databasename
    SET PORT=port on which the web-server listens
    SET MAIL_HOST=smtp-server hostname
    SET MAIL_PORT=smtp-server port number (usually 25)
    SET MAIL_LOGIN=login of smtp-server account
    SET MAIL_PASSWORD=password of smtp-server account
    SET MAIL_FROM=E-mail address, from which will be sent emails

# WEB-server API

## `GET /`
Returns main html-page. The page is returned empty, without working information, such as tasks, lists or current user. Only list of languages and current-language value are included. Instead, the page contains javascript for authorization and further work with tasks.

## `GET /static/...`
Returns static files: icons, javascripts libs, css stylesheets, etc.

## `POST /SignUp`
Try to sign-up a new user. 
In case of success, a link is sent to the user, after which he can set password and complete the registration. 
Without opening the link, the account is not valid.

	IN: JSON: { email : "" }
	OUT: JSON: { result : string ["EMailEmpty", "UserJustExistsButEmailSent", "UserSignedUpEmailSent"] }

# Database structure

## `Tasks`

Main collection, that contains task records.

## `Users`

Collection of users and their hashed passwords.

# How we name variables in Go?

We use both: lowerCamelCase or UpperCamelCase:
* If the variable is global and should be visible in more than one procedure, then use **UpperCamelCase**.
* For local variables that are used on short segments of the code, we use **lowerCamelCase**.
* For **type identifiers**, we use lowerCamelCase, which begins with the word "type": eg `typeTaskList`.
* All **fields of structures** are named in UpperCamelCase.
* We use **short names** of a small number of letters (eg `i`, `err`) if they are used in no more than several neighboring lines of code.

# How we name classes and id's in HTML DOM?

We use **lower-case-with-dash-delimeter** style. If there are several elements with these characteristics on the page, give them a common **class** name. If this element is unique - specify its **id**. Elements with interactivity and with which the javascript interacts, must necessarily have defined id's. For some elements there are both: class and id defined.

This is not good, if the **class-name** contains the tag-name to which this class applies. A class with the same name can have different CSS-implementations for different elements of the HTML DOM. Good idea for class name is `task` or `icon`. Bad idea is `task-div` or `icon-img`. 
Opposite the **id** should contain the name of the element's tag, for example: `edit-task-form` or `operation-status-label`.

If the class name contains its distinctive qualities (modifiers), then they should be written via a double dash: `<p class="task--done">`. If the modifier is global and can be applied not only to this class, it should be allocated to a separate stand-alone class: `<p class="task done">`...`<img class="done">`.

The name of class and id's should reflect the content and purpose of the element, not its location or characteristics. Good idea for class name is `task--done`, bad idea is `green-strikeout`. Goog idea is to use CSS for aling div-tag with id `search-div`. It's a bad idea to give him class `right-align`.

# Other notes

After changing files `tasks.css` and `tasks.js`, increase their suffix in `<link>` and `<script>` tags `tasks.html` template. Browsers, that have this files cached, will apply the changes the next time the page is reloaded.