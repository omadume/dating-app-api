# Dating App API

## Build and Run the app in a Docker Container

Open a terminal application and run the following commands.

Build the Docker image:

`docker build -t dating-app-api .`

Run the Docker container:

`docker run -p 8080:8080 --rm dating-app-api`

Now the app should be running and available at http://localhost:8080.

## Testing the Endpoints (examples)

Open the docker container terminal and run the following commands.

### Create a User

To create a user (this will return a randomly-generated user currently):

`curl -X POST http://localhost:8080/user/create`

### Login

To login with an email and password (this will return an authentication token):

`curl -X POST -d '{"email":"user@example.com", "password":"password"}' http://localhost:8080/login`

### Discover Profiles

To fetch profiles of potential matches (this will return all registered users currently):

`curl -H "Authorization: Bearer <token>" http://localhost:8080/discover`
- Replace `<token>` with the JWT token received from the login endpoint.

Optionally, you can also filter by age (18-90), gender (female, male, or z) or distance (in km):

`curl -H "Authorization: Bearer <token>" "http://localhost:8080/discover?min_age=25&max_age=35"`

`curl -H "Authorization: Bearer <token>" "http://localhost:8080/discover?gender=female"`

`curl -H "Authorization: Bearer <token>" "http://localhost:8080/discover?max_distance=10"`
- Note: Make sure to wrap the URL in quotes ("") to ensure the correct parsing of all query parameters

You can also combine filters:

`curl -H "Authorization: Bearer <token>" "http://localhost:8080/discover?min_age=25&max_age=35&gender=female"`

### Swipe

To respond to a profile with a preference (YES or NO):

`curl -X POST -H "Authorization: Bearer <token>" -d '{"targetUserId":1, "preference":"YES"}' http://localhost:8080/swipe`
- Replace `<token>` with the JWT token received from the login endpoint.

---

### Assumptions & Decisions

- Assumption that json response fields do not need to adhere to any specific order.
- Assumption that user auth token is not persisted as there is no client-side to this app.
- Decision to allow the same user to login multiple times and receive a new auth token each time (even if previous one has not yet expired - any existing token is overriden).
- Decision to leave all files at the root folder as the app is a simple one - could be refactored into subfolders upon expansion.