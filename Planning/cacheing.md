so basicaly there are 2 options 

I dont want to just have a button to get the entire cache, i want to like be able to fetch the cache dynamicaly, based on user needs.

The reason for a cache is mainly for offline usage of the app (mobile app).

## Copy entire DB / required fields

This would be advantagouns, because i can write custom queries in the frontend based on the chached data and tthru this have direct access to ofline data without missynced data.

An issue i would see is that i dont have up to date data or if so then I would have to do a lot of inserts, which i dont want to.

Maybe this isnt even that bad, beacuse i am able to query the "dubplicated" db directly. An issue i see coming up would be with deleted data. I think checks foreach group / meal / user and if they still exist would be pretty painful and difficult to implement and also kinda expensive on client load. 

### Pros: 
- Its incredibly robust
- Wont have any Overlapping cache invalidities
- Will be easie to expand and build uppon

### Cons: 
- issue with deleted data
- Difficult to implement and not feasable for this project size. (would be realy usefull on a bigger app)

## Only a cache key value db

this would have the advantage of beeing way easier and simpler to handle, because i would just try to hit the cached data if a user wants to request a specific route. For example if the user wants to access the groupData with the id `1` the route would look like this: GET: `groups/1`. now this could be a unique Cache key, from which is first being checkd if it exists, and if yes, it would reuturn the cached data instead of hitting the db. 

There would also a datetime like 30 seconds or something into the future called `isInvalidCache`. when this is reached the server would be queried, and if that fails too it would show the invalid data (this would be the case for offline usage).

An Issue i see with that is if i for example cache the data for `group/1` and then also cache the data for a specifc meal in this group, there would be data inconsistencies.
The MealCard in the group would for example still display `X` amount of opted in people, but the meal display would show `N` amount of people.
The problem happens, becuase the data displayed in the mealCard is being fetched seperatly, than the data fetched for the entire meal data.

I call this issue not a cahce invalidity, but a cache missync, because 2 differing data blobs are cached, and both dont know of the existance of the other. 

### Pros: 
- Simpler to implement 
- Cache validity is almost every single cache entry given, if its being requested while online
- Generaly just simple

### Cons: 
- Issue with overlaping data inconsistency
- difficult to expand (wont be a problem here, because the app will not become insaley big)


[ChatGPT discussion](https://chatgpt.com/share/673f6ae7-f3d8-8005-a001-565a912fb827)