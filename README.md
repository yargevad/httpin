# httpin
net/http compatible middleware for automatic parsing of http input

## What is this?
This middleware parses http input using types defined as part of router setup.
It stores the resulting struct pointer into the request context, along with an embedded pointer to app state.
