### How to run
docker-compose up --build

### Endpoints
#   http://localhost:3000/users

GET: Gets all users
POST: Creates new user with fields: name, password, role (admin or user). When new user is created, his token is returned.
This token is needed to add, delete and patch companies as authentication. Only users with admin role can implement these functions

# http://localhost:3000/company
GET: Gets all companies
POST: Creates new company. On headers it should be added for authentication pair(   Key: "x-jwt-token", Value: [Token retrived before]).
If token matches admin user, permission to complete create will be granted. If user is created with role user, he will not have access to modify companies

# http://localhost:3000/company/{id}
DELETE
PATCH
Same jwt process as above is needed

### Example 

1. http://localhost:3000/users
    Create new POST request, example body:      
    {
        "username": "katerina"
        "password": "password",
        "role": "admin"
    }
    Keep the token returned upon retrieval for example "example-token"

2.  http://localhost:3000/company
    Create new POST request, example body:      
    {
        "name": "google",
        "description": "search engine",
        "employeesAmount": 100,
        "registered": false,
        "companyType": "Corporations"
    }
    On headers insert new pair 
    {
        Key: "x-jwt-token",
        Value: "example-token"
    } 

    Events are produced with message loging with every database mutation