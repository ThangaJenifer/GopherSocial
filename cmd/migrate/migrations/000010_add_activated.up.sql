ALTER TABLE
    users
ADD
    COLUMN is_active BOOLEAN NOT NULL DEFAULT FALSE;

/*

Now just one more thing here on the database which is a table or not a table, a column that we need  
to create which is the activated on the user ID, you can see that we don't have any way to know if this 
user is activated or not. So what I have just done here very quickly is that I created a new migration code and activated the
user, and I just altered the users table to add a new boolean, just say is active and the default value 
is going to be false.So all of the users that we have created are false.If, of course, none of those users existed.So I don't mind having this false because we're going to delete all of these users once we deploy to
production, or actually we're not going to delete them because it's going to be a new database on production.
So there's going to be no users and the defaults needs to be false.

*/