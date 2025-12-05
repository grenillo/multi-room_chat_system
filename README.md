How to deploy:

    1. To use the multi-room chat system you must first ensure your terminal is in the ./multi-room_chat_system/main

    2. From here to start the server state, into your terminal you must enter:
                go run ./server

    3. To start any client, while you are still in ./multi-room_chat_system/main, in any number of terminals (for any number of clients) enter:
                go run ./client
        
    4. To quit any client while the server is still running, in the client GUI, enter:
                /quit
    
    5. To shutdown and save the state of the server, in the OWNER GUI, enter:
                /shutdown


Notes:
    1. There are 3 roles preloaded onto the server:
            a. user -> has role Member
            b. admin -> has role Admin
            c. owner -> has role Owner
                The owner is the ONLY user who can shutdown the server
        all other users can be added to the state at runtime

    2. There are 2 preloaded rooms on this server:
            a. #general -> room for ALL users
            b. #staff -> room for users with Admin or Owner roles

    3. Ensure server is running before starting any clients
