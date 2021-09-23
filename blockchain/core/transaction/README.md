
##  Struct of module

#### DDD comprises of 4 Layers:

>**_Domain_:** This is where the domain and business logic of the application is defined.

> ~~**_Infrastructure:_** This layer consists of everything that exists independently of our application: external libraries, database engines, and so on.~~
 
>**_Services:_** This layer serves as a passage between the domain and the interface layer. The sends the requests from the interface layer to the domain layer, which processes it and returns a response.
    
>**_API:_** aka _Controller_ aka _Interface_. This layer holds everything that interacts with other  modules, such as block or state and so on.
