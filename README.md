# dwr
Diminishing Weighted Random Distribution implementation, with state save

# Input Register json
POST /myKey JSON:
json := {
                "weight_1" : 1,
                "weight_2" : 2,
                "weight_heavy": 99
        }

# Query
Request : GET <host>/myKey

# Response
{ "key": "weight_heavy" }

# DELETE
Request: DELETE <host>/myKey

    
    



