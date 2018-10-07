# The task list for easy work with a large number of tasks (**100+**), which makes it easy to start regularly **from scratch**

Experimental web-app ToDo-list by Stanislav Kornienko's conception. You can try it in the cloud <https://todo.works>

<img src="https://github.com/st-korn/KornToDoList/raw/master/static/favicon.png" width="64">

## Project Files

* `main.go` - main part of backend server application, that implements common functions, global variables and constants, Go-types of database collections, starts web-server with HTTP static-files transfer.
* `web.go` - implements interfaces for HTML-generation
* `users.go` - implements JSON-api of users-routines (such as SignIn, LogIn, LogOut, etc..) for javascript and mobile applications.
* `lists.go` - implements JSON-api of lists-routines (such as GetLists, CreateList) for javascript and mobile applications.
* `tasks.go` - implements JSON-api of tasks-routines (such as GetTasks, SendTask) for javascript and mobile applications.
* `templates/` - here are HTML-templates and JSON language packages. A file from this folder is used by the main server application, but not transmitted to clients directly via HTTP requests.
* `static/` - here are the static files that clients should receive via http requests: javascripts, CSSs, icons, etc.
* `.graphics/` - in this folder the original images and the results of their processing are saved. All images received from third-party sites and services must necessarily contain a license, which allows their public commercial use.
* `.vscode/` - here the Visual Studio Code settings are saved.

## Environment variables

To run .go server-application you need to set these environment variables:

    SET MONGODB_ADDON_URI=mongodb://username:password@domain.com:port/databasename
    SET MONGODB_ADDON_DB=databasename
    SET MAIL_HOST=smtp-server hostname (eg. smtp.mail.ru)
    SET MAIL_PORT=smtp-server port number (usually 25)
    SET MAIL_LOGIN=login of smtp-server account
    SET MAIL_PASSWORD=password of smtp-server account
    SET MAIL_FROM=E-mail address, from which will be sent emails
    SET SERVER_HTTP_ADDRESS=http-server hostname with http or https prefix, and without any path. User in email sent. (eg. <https://todo.works>)
    SET LISTEN_PORT=port on which the web-server listens HTTP-requests (for example 8080 for golang applications, hosted on clever-cloud.com)

## WEB-server conception

* The application accepts incoming only **HTTP**-requests on the port `LISTEN_PORT`. The application does not listen or open any HTTPS connections.

* Internal cloud HTTPS-proxy servers usually take over SSL encryption and translate to the application only HTTP-requests inside the cloud infrastructure. For example, <https://www.clever-cloud.com>. When translating requests, a cloud proxy usually sets a header `X-Forwarded-Proto`. Using this header, you can determine how the initial request from the user's browser arrived: over HTTP or HTTPS. <https://www.clever-cloud.com/blog/engineering/2015/12/01/redirect-to-https-in-play/>

* If the request from the browser came over the HTTP-protocol, the system returns the `301 (Moved Permanently)` redirection to the corresponding HTTPS page.

* The application allows the receipt of requests directly without a cloud proxy (for example, for debugging). The application processes such requests only via the **HTTP**-protocol, and they are not redirected to HTTPS.

## WEB-server API

### `GET /`, `GET /YYYY-MM-DD`

Returns main html-page. The page is returned empty, without working information, such as tasks, lists or current user. Only list of languages and current-language value are included. Instead, the page contains javascript for authorization and further work with tasks.

    Cookies: User-Language : string

If the name of the list in the format `YYYY-MM-DD` is present in the URL, then if the authorization is successful, the web application tries to open the list with that name. If the authorized user does not have a list with the same name, or the name of the list is not specified, the most recent uf users list opens.

### `GET /static/...`

Returns static files: icons, javascripts libs, css stylesheets, etc.

### `POST /SignUp`

Try to sign-up a new user.
In case of success, a link is sent to the user, after which he can set password and complete the registration.
Without opening the link, the account is not valid.

    IN: JSON: { EMail : string }
    OUT: JSON: { Result : string ["EmptyEMail", "UserJustExistsButEmailSent", "UserSignedUpAndEmailSent"] }

### `GET /ChangePassword?uuid=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`

Returns html-page to set users-password.

    Cookies: User-Language : string
    IN: uuid : string

* Then find uuid parameter in `SetPasswordLinks` table.
* If not found - returns html-page with errors and "Return" button.
* If found -returns html-page with two input boxes for new password.
* When html-form is submited, it calls POST /SetPassword

### `POST /SetPassword`

Add new user with password or change password for existing user.

    IN: JSON: { UUID : string, PasswordMD5 : string }
    OUT: JSON: { Result : string ["UUIDExpiredOrNotFound", "EmptyPassword", "UserCreated", "PasswordUpdated"] }

* Try to find current set-password-link. If not found - return `"UUIDExpiredOrNotFound"`
* If access is allowed, insert a document in MongoDB `Users` collection, or update it.
* If success delete UUID record from MongoDB `SetPasswordLinks` collection (to block access to a link)
* Then return `"UserCreated"` or `"PasswordUpdated"`.

### `POST /LogIn`

Try to login an user, returns session-UUID.

    Cookies: User-Session : string (UUID)
    IN: JSON: { EMail : string, PasswordMD5 : string }
    OUT: JSON: { Result : string ["EmptyEMail", "EmptyPassword", "UserAndPasswordPairNotFound", "LoggedIn"], UUID : string }

* If current session exist - then logout.
* If the pair (EMail and PasswordMD5) is not present in the collection `Users`, return "UserAndPasswordPairNotFound".
* Otherwise register new session in the `Sessions` database collection, and return its UUID to set cookie in browser.

### `POST /LogOut`

Logout user and erase information in database about his active session.

    Cookies: User-Session : string (UUID)
    IN: -
    OUT: JSON: { Result : string ["EmptySession", "LoggedOut"] }

### `POST /GoAnonymous`

Start an anonymous-session, returns session-UUID.

    Cookies: User-Session : string (UUID)
    IN: -
    OUT: JSON: { Result : string ["SuccessAnonymous"], UUID : string }

* If current session exist - then logout.
* Register new anonymous session in the `Sessions` database collection, and return its UUID to set cookie in browser.

### `POST /UserInfo`

Check session's UUID and get information of current user.

    Cookies: User-Session : string (UUID)
    IN: -
    OUT: JSON: { Result : string ["ValidUserSession", "ValidAnonymousSession", "SessionEmptyNotFoundOrExpired"], EMail : string }

* Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
* Returns the Email (real or imaginary) of the current user. Returns flag: is the current user anonymous or not.

### `POST /GetLists`

Get list of user lists.

    Cookies: User-Session : string (UUID)
    IN: -
    OUT: JSON: { Result : string ["OK", "SessionEmptyNotFoundOrExpired"], Lists : []string }

* Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
* Returns array of strings with names of saved users todo-lists.

### `POST /GetTasks`

Get tasks of selected user lists.

    Cookies: User-Session : string (UUID)
    IN: JSON: {List : string}
    OUT: JSON: { Result : string ["OK", "SessionEmptyNotFoundOrExpired"], Tasks : [] { Id : string, EMail : string, List : string,
    Text : string, Section : string ["iu","in","nu","nn","ib"], status : string ["created", "done", "canceled", "moved"],
    Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"], Timestamp : datetime } }

* Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
* Returns an array of structures that identify tasks from a selected list of the current user.

### `POST /SendTask`

Update existing task from the list or append new task to the list.

    Cookies: User-Session : string (UUID)
    IN: JSON: { List : string, Id : string (may be null or ""), Text : string,
    Section : string ["iu","in","nu","nn","ib"], Status : string ["created", "done", "canceled", "moved"],
    Icon : string ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"], Timestamp : datetime }
    OUT: JSON: { Result : string ["TaskEmpty", "InvalidListName", "SessionEmptyNotFoundOrExpired", "UpdatedTaskNotFound", "UpdateFailed", "TaskJustUpdated", "TaskUpdated", "InsertFailed", "TaskInserted"],
    Tasks : [] { Id : string, EMail : string, List : string, Text : string, Section : string, Status : string, Icon : string, Timestamp : datetime } }

* Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
* If updated task exist in database, and its timestamp is greater than timestamp of updated task, recieved from users-application, return `"TaskJustUpdated"` error.
* Update existing task or generate ID and append new task to the database.
* Returns an array of a single element - an added or updated task with its ID.

### `POST /CreateList`

Create new todo-list for current user.

    Cookies: User-Session : string (UUID)
    IN: JSON: {List : string "YYYY-MM-DD"}
    OUT: JSON: { Result : string ["ListCreated", "InvalidListName", "DateTooFar", "CreateListFailed", "SessionEmptyNotFoundOrExpired"], Lists : []string }

* Checks the current session for validity. If the session is not valid, it returns `"SessionEmptyNotFoundOrExpired"` as a result.
* Verifies that the desired list name does not differ by more than 24 hours from the current server date.
* Returns array of strings with names of saved users todo-lists.

## Database structure

### `Tasks`

Main collection, that contains task records.

    {
        "_id" : ObjectId // unique object identifier
        "email" : string // username in format "user@domain.net" for registred users or "IP@YYYY-MM-DD-hh-mm-ss:YYYY-MM-DD" for anonymous user.
        "list" : string // list name in format: "YYYY-MM-DD"
        "text" : string // tasks title, for example "Peter - send invoice"
        "section" : string // the importance and urgency of the task: ["iu","in","nu","nn","ib"]
        "icon" : string // one of the icons: ["wait","remind","call","force","mail","prepare","manage","meet","visit","make","journey","think"]
        "status" : string // task status: ["created", "done", "canceled", "moved"]
        "timestamp" : datetime // server date and time last modification of task
    }

### `SetPasswordLinks`

Collection, that contains links for changing user passwords.

    {
        "email" : string // email address for password change
        "uuid" : string // UUID to access password changing
        "expired" : datetime UTC // UTC-datetime to which the link will be valid
    }

### `Users`

Collection, that contains registered users records and their hashed passwords.

    {
        "email" : string // email address of the registred user
        "passwordmd5" : string // MD5-hash of his password
    }

### `Sessions`

Collection, that contains active cookies sessions for users.

    {
        "uuid" : string // UUID cookie of the session
        "email" : string // email address of the session user
        "expired" : datetime UTC // UTC-datetime to which the session is valid
    }

## Identifiers naming

### How we name identifiers in Go

We use both: lowerCamelCase or UpperCamelCase:

* If the variable is global and should be visible in more than one procedure, then use **UpperCamelCase**.
* For local variables that are used on short segments of the code, we use **lowerCamelCase**.
* Global variables, which are set once at the start of the program and used in most functions, are written **UPPERCASE**.
* For **type identifiers**, we use lowerCamelCase, which begins with the word "type": eg `typeTaskList`.
* All **fields of structures** are named in UpperCamelCase.
* All **function parameters** must be lowerCamelCase.
* We use **short names** of a small number of letters (eg `i`, `err`) if they are used in no more than several neighboring lines of code.
* It's a good way to comment: why every external package imported in the program.

### How we name identifiers in JavaScript

* Function and variables in JavaScript are named in **lowerCamelCase**.
* All JSON filed names must be **UpperCamelCase**, same as fields of structures in Go. Do not forget to use correct case of JSON filed names in Javascript ajax-routines.
* All web API-calls must be **UpperCamelCase** too.

### How we name identifiers in MongoDB

* MongoDB collections are **UpperCamelCase** (eg `SetPasswordLinks`)
* All MongoDB field names must be **lowercase**, in contrast to the fields of structures in Go. Do not forget to use correct case of collection fileds in all Go-routines (eg. `passwordmd5`).
* EMail addresses must be stored in database only in **lowercase** style also. It's a good way to lowercase incomming email address in every backend Go-routine.

### How we name classes and id's in HTML DOM

We use **lower-case-with-dash-delimeter** style.

* If there are several elements with these characteristics on the page, give them a common **class** name. If this element is unique - specify its **id**. Elements with interactivity and with which the javascript interacts, must necessarily have defined id's. For some elements there are both: class and id defined.
* If you need to access  an element from JavaScript or set unique characteristics of an element in CSS, determine and use the element **ID**.
* If similar characteristics and behavior should be on several elements with the same tag, but not on all, define and use a **class**.
* If all elements of this tag must have the same behavior and characteristics - you can work directly with the **tag**, without classes or ID.
* This is not good, if the **class-name** contains the tag-name to which this class applies. A class with the same name can have different CSS-implementations for different elements of the HTML DOM. Good idea for class name is `task` or `icon`. Bad idea is `task-div` or `icon-img`.
* Opposite the **id** should contain the name of the element's tag, for example: `edit-task-form` or `operation-status-label`.
* If the class name contains its distinctive qualities (modifiers), then they should be written via a double dash: `<p class="task--done">`. If the modifier is global and can be applied not only to this class, it should be allocated to a separate stand-alone class: `<p class="task done">`...`<img class="done">`.
* The name of class and id's should reflect the content and purpose of the element, not its location or characteristics. Good idea for class name is `task-done`, bad idea is `green-strikeout`. Goog idea is to use CSS for aling div-tag with id `search-div`. It's a bad idea to give him class `right-align`.
* To determine the CSS properties of objects of a given class or ID, do not refer to their tags. It is wrong to write `div#welcome-div`, right `#welcome-div`.
* It is not necessary to start each html tag from a new line. You can fit several html tags in one line, if this improves the readability of the document.

## Other notes

After changing files `.css` and `.js`, you can increase their suffix in `<link>` and `<script>` tags `.html` template. Browsers, that have this files cached, will apply the changes the next time the page is reloaded.