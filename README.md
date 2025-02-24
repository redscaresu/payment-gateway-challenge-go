# Instructions for candidates

This is the Go version of the Payment Gateway challenge. If you haven't already read the [README.md](https://github.com/cko-recruitment/) in the root of this organisation, please do so now. 

## Template structure
```
main.go - a skeleton Payment Gateway API
imposters/ - contains the bank simulator configuration. Don't change this
docs/docs.go - Generated file by Swaggo
.editorconfig - don't change this. It ensures a consistent set of rules for submissions when reformatting code
docker-compose.yml - configures the bank simulator
.goreleaser.yml - Goreleaser configuration
```

Feel free to change the structure of the solution, use a different test library etc.

### Swagger
This template uses Swaggo to autodocument the API and create a Swagger spec. The Swagger UI is available at http://localhost:8090/swagger/index.html.

### Demo Playbook

run the application in debug mode via vscode and run docker compose up

#### Happy Path PostPayment authorized
```
curl -X POST http://localhost:8090/api/payments \
-H "Content-Type: application/json" \
-d '{
  "card_number": 2222405343248877,  
  "expiry_month": 4,
  "expiry_year": 2025,
  "currency": "GBP",
  "amount": 100,
  "cvv": 123
}' | jq .
```

#### Happy path Get Authorized Payment
```
curl -X GET http://localhost:8090/api/payments/$id | jq .
```
#### Unhappy path Get Payment does not exist
```
curl -vvvv -X GET http://localhost:8090/api/payments/foo | jq .
```
#### Unhappy Path declined
```
curl -X POST http://localhost:8090/api/payments \
-H "Content-Type: application/json" \
-d '{
  "card_number": 2222405343248878,  
  "expiry_month": 4,
  "expiry_year": 2025,
  "currency": "GBP",
  "amount": 100,
  "cvv": 123
}' | jq .
```
#### Unhappy path Get Payment Declined
```
curl -vvvv -X GET http://localhost:8090/api/payments/$id | jq .
```
#### Unhappy Path rejected (incorrect card number)
```
curl -X POST http://localhost:8090/api/payments \
-H "Content-Type: application/json" \
-d '{
  "card_number": 1,               
  "expiry_month": 4,
  "expiry_year": 2025,
  "currency": "GBP",
  "amount": 100,
  "cvv": 123
}' | jq .
```
#### Unhappy Path upstream 503 from acquiring bank
```
curl -X POST http://localhost:8090/api/payments \
-H "Content-Type: application/json" \
-d '{
  "card_number": 2222405343248870,  
  "expiry_month": 4,
  "expiry_year": 2025,
  "currency": "GBP",
  "amount": 100,
  "cvv": 123
}' | jq .
```
### Solution Commentary

My solution creates a set of handlers and corresponding domain methods alongside a client.  The domain and client are mockable so as to be able to test each tier of the application in isolation, I also include some integration tests using mountebank.  Please note that mountebank needs to be running with a docker compose up before running the integration tests.

#### Integration tests

Integration tests use the the mountebank docker container which will potentially take longer to run in a pipeline so I tested the main unhappy paths and happy paths but there is an argument to say we should aim for more test coverage via the integration tests because it is testing the real code.

In the case of the integration tests I tested 1 validation, 503 failure with the acquiring bank and also the happy POST and GET on a payment.  Given more time, I would test all of the validations.  

TODO: I would like to create the container within the test.  In order to do this we would need to do some light refactoring and inject the url of the mounteabank into the api.New() in api.go.  We could then interact directly with the docker container libraries to spin up and down our container for each test.

#### Handlers Implementation approach

For the handlers implementation I split away as much of the business logic into the domain tier to keep the handlers as clean as possible.  Also for the case of the GET I did a direct call to the storage layer from the handler.  If we need more complex logic in time I would eventually shift it into the domain but for the purposes of YAGNI for the time being only the POST has a corresponding domain method.  The post does contain some more complicated logic so for the purposes of cleanliness I split out the code into the domain.

The main thing the handlers do is check whether there is an error being returned or not from the domain and convert it into a public error.

TODO: split out the error handling on the handlers into its own package, test them in isolation.

TODO: discuss with product manager error shapes, with validation errors we could return the field the customer errored on, this is already logged on our system to help with future app support queries.

#### Handlers Test approach

For the tests I pretty much left the GET method as it was and tested the main paths.
However for the POST tests I did not reuse the t.Run() style. I find the t.Run() a bit hard to read and I have found issues with maintainability once you start trying to manage state against lots of tests.  So for this reason each test has its own state separate from one another.

For the integration with the domain I inject the payments service, this is so that I can mock exactly the output I want from the payments service into the handler.

TODO: work on some test helpers to clean the test code.

TODO: review use of mocking, can we just inject mountebank directly into the client without mocking the payservice?

#### Domain Implementation approach

I create a Domain that we are able to inject payment services into, the reason why I did this was so that we could inject mocks.

Arguable YAGNI but I created a domain that can contain n services, for example if end up having to include a new service to integrate with in the future its just a simple case of adding an additional service to the Domain struct and creating the requisitive methods and interfaces.  It also allows us to split the implementation away from the interface of the services here.

#### Domain Testing approach

For the domain testing we mock the client, which allows us greater flexbility to test for all possible responses from the client we are integrating with.  We test all the validations and possible return values from the client to the domain.

TODO: make the validation testing more exhaustive for example on the card number we only test that the card number is too short not that it is too long.  Could be a good candidate for some table tests here.

TODO: review use of mocks here, they could be replaced by mountebank.  Where mountebank does not support a response from the real acquiring bank we could extend the functionality of it to include the new response

#### Client Implementation Approach

I include an interface so that we can mock the client and test for all possible responses from the bank.

TODO: Make the client generic, we could have a method called "DO" and then pass in the verb and url from the domain so we dont have create new methods for a new endpoint.

#### Client Test Approach

Using the testserver to fake responses from our acquiring bank and asserting that the errors are correctly handled.

TODO: Review use of mocks here, potential to use mountebank and extend it where necessary.

TODO: Greater test coverage on all the paths, the tests here are not exhaustive.