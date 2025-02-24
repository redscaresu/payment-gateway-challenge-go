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

run the application in debug mode via vscode.

#### Happy Path PostPayment authorized

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

#### happy path Get Authorized Payment

curl -X GET http://localhost:8090/api/payments/$id | jq .

#### unhappy path Get Payment does not exist

curl -vvvv -X GET http://localhost:8090/api/payments/foo | jq .

#### Unhappy Path declined

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

#### unhappy path Get Payment Declined

curl -vvvv -X GET http://localhost:8090/api/payments/$id | jq .

#### Unhappy Path rejected (incorrect card number)

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

#### Unhappy Path upstream 503 from acquiring bank

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

### Solution Commentary

#### Integration tests

All of these main unhappy and happy paths are tested via integration held in /integration.

Integration tests use the the fake acquiring bank as a docker container rather than mocks, they take longer to run so for the purposes of a quick pipeline I tested the main unhappy paths and happy paths but there is an argument to say we aim for more test coverage via the integration tests because its testing the real application.  Also because I want a TDD approach it is much easier to write code with an accompanying unit test and then layering on an integration test once the main code structures are there.

In the case of the integration tests I tested 1 validation, 503 failure with the acquiring bank and also the happy POST and GET on a payment.  I had more time I would of tested all of the validations.  If I had more time I would configure the mountebank to start in the background of a integration test without having to manually start the container via docker compose.  In order to do this we would need to do some light refactoring and inject the url of the mounteabank into the api.New() in api.go we could then interact directly with the docker container libraries to spin up and down our container for each test.

#### Handlers approach

For the tests I pretty much left the Get method as it was and tested the main paths.
However for the POST tests I did not reuse the t.Run() style I find the t.Run() a bit hard to read and I have found issues with maintainability once you start trying to manage state against lots of tests so for this reason each test has its own state separate from one another however this has resulted in a lot of duplicated code.

For the integration with the domain I inject the payments service this is so that I can mock exactly the output I want from the payments service into the handler.  If I had more time I would work on some test helpers to make this cleaner as the set up is repeated in lots of tests or implementing an wrapper.  I am not sure if mocking was the right way to go here, maybe I should of just injected the mountebank client into the paymentservice so that the domain is real but the client is using the fake or a mocked client.

For the handlers implementation I split away as much of the business logic into the domain tier to keep the handlers as clean as possible.  Also for the case of the get I did a direct call to the storage layer from the handler, if we need more complex logic in time I would eventually shift it into the domain but for the purposes of YAGNI for the time being only the POST has a corresponding domain method.  The post does contain some more complicated logic so for the purposes of cleanliness I split out the code into the domain.

The main thing the handlers do is check whether there is an error being returned or not and if there is make sure that the error is correctly handled.  If I had more time I would refactor this out and potentially create a package that matched the application errors to the corresponding public error that we wanted to show.

#### Domain approach

I create a Domain that we are able to inject payment services into, the reason why I did this was so that we could inject mocked objects into the payment service.  These are not as strong as the integration tests but run faster.

Arguable YAGNI but I created a domain that can contain n services, for example if end up having to include a new service to integrate with in the future its just a simple case of adding an additional service to the Domain struct and creating the requisitive methods and interfaces.  It also allows us to split the implementation away from the interface of the services here.

If I was to do this again I am not sure I would mock this, I would probably use the mountebank container instead and use the real domain.  I would set up the client so that it is injectable from main.go

#### Client Approach

I include an interface so that we can mock the client but thinking about it I am not sure that is necessary we could have used the fake bank but at least now it means we can test all possible responses from the bank including errors like a 500.  I could of made this more generic by having a method called "DO" and then we could have passed in the method and the URL that way in future as we add more endpoints we can just call it with a different path as needed.  As it stands a new method would need to be created for each endpoint.

In terms of the tests at the moment I am using the testserver but I could of used the fake bank to test the client instead.  But I still think this is a good way to test the client as it is a simple client and we can test all the possible responses from the bank.  The tests are not exhaustive here if i had more time I would test all the possible responses from the bank.