# dwr
Diminishing Weighted Random Distribution implementation, with state save

# Input Register json
    
json := {
            "key" : "myKey",
            "weights" : {
                "weight_1" : 1,
                "weight_2" : 2,
                "weight_heavy": 99
            }
        }

# Query
Request : GET <host>?key="myKey"

# Response
{ "value": "weight_heavy" }

    
    



