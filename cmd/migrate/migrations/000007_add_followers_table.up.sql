CREATE TABLE IF NOT EXISTS followers (
    user_id bigint NOT NULL,
    follower_id bigint NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, follower_id), --composite key
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE, -- foreign key referencing user's ID
    FOREIGN KEY (follower_id) REFERENCES users (id) ON DELETE CASCADE -- foreign key referencing follower
);

/*
Now this is going to be important because the primary key is going to be a user and the follower ID. Why?
Because you don't want to have duplicate entries. Imagine user one starts following user two.
Now user one can follow two. User three can follow user four. User five and so on and so on.
But the thing is, user one cannot follow user two again. So we don't want to have this duplication of data. 
What I mean is that we should not allow this behavior to go unnoticed.

So let's do a composite key.  And this is how you do composite keys.  This is this is possible.  
If you're not familiar with this you can just do user ID and then a follower ID.  
And this is going to be considered the primary key.  So this is called composite key okay.

Then let's add also the foreign keys because this this is also a foreign key which references the user's  ID.  
We have already done this.  So um I'm going to have maybe delete on Cascades.  This is not important data that I want to keep if the user is deleted.  
So I'm fine with this.  Usually I'm very um I'm doubtful on using delete cascade because most of the code bases are work there.  Soft delete.  
For example, for entities like users, he wants to soft delete instead of doing a hard delete like  we're currently doing.  Because imagine a client deletes a user and we want to rollback.  
We can actually do it with a soft delete, because although the user information is going to be deleted,  we can revert it because we know the user by the logs and all of that.  But the important thing we keep is the relations which the users and any metadata the user might have  on the on the database.  
But I know for a fact that this is going to be an application for a consumer, not a business.  So if the user deletes his or her account, I'm not that concerned that we are going to revert it back.  So I'm going to go with this decision.  If you think you should add soft to the leads, just take in mind that there's on delete.  
Cascade is not a good idea for that.
*/