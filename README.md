# mailbridge

REST to Email Bridge for use in Web Forms

## description ##

This is a simple application for use in web forms to send emails to predefined addresses, eg. in contact forms.

It exposes two endpoints:

* GET /api/token

    Get a token. This token is stored in memory and must be provided in the next request to actually send emails. It will be deleted after the first usage.
    If a user requests another token after short time, this endpoint will slow down the response to make spammers lifes harder. see **Tarpit** below
    
* POST /api/send

    Send an email.
    This will need a JSON body with following fields filled in:
    
    <pre>
    {
      "From":"Sender email address", 
      "To": "Recipient Identifier as defined in Config", 
      "Subject": "Subject of the mail", 
      "Body": "Mail Body", 
      "Token": "the Token as received from the token endpoint"
    }
    </pre>

## config ##

you will need a configuration file like the following:

<pre>
  "host": "mail.example.com",
  "port": "25",
  "authUser": "SMTP_USER",
  "authPassword": "SMTP_PASSWORD",
  "recipients": {
    "id1": "one_email@example.com",
    "id2": "another_email@example.com"
  },
  "lifetime": 60,
  "cleanupInterval": 10,
  "tarpitInterval" : 10</pre>

* host: the SMTP host to connect to
* port: the SMTP port to be used
* authUser: the Username part of the SMTP Authentication
* authPassword: Password for the SMTP authentication
* recipients: Map of arbitrary IDs to email addresses. The form will need to provide the defined ID, the application replaces that by the referring email address
* lifetime: lifetime of a token in seconds, after this time a new token will expire
* cleanupInterval: interval in seconds how often a cleanup run will delete expired tokens
* tarpitInterval: interval in seconds for how long a user token request should be delayed if the user sent already requests short time ago 

## Tarpit ##

The token endpoint will store the IP Address of the client in memory for a short period of time, as defined in **tarpitInterval** in the configuration.

If the same client sends another request during this time, the token response will wait for this interval before answering.
The number of requests during that interval is incremented, so if a client sends the third request, the application will wait 3 times the tarpitInterval before answering.

## Status ##

This is not yet ready to use, so pre-alpha I would say.

## TODO ##

* write tests for mail sending, controller and for the tarpit
* write docker file and create CI jobs 
* add authenticated metrics endpoint with version, timestamp and counter values for requests, errors and sent mails, runtime for cleanup goroutimes

## License ##

This application is licensed under GPL3, see the file **LICENSE** for details.

## Author ##

Alexander Köb <github@koeb.me>
