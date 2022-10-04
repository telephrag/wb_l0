# NOTES
01. Tech stack:
     -- nats-streaming      message passing to service
     -- postgresql          persistant storage
     -- bbolt               in-memory caching
02. If received JSON has more fields than model excessive fields will be dropped.

# TODO
01. Create generator of random objects of type `Order`. Make `publisher` stream them along 
    with some junk data.
02. Implement validation to accept only complete payloads.
