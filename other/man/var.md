USAGE:  [.var:|var][VAR_NAME] = [VALUE]
Assigns a value to a variable   

Variable names are solely alphanumeric characters   
With the first character being a letter   
Assignable values are bool, int, string, array, function, node, json or path.   

Path variables without quotes can only be assigned using the 'var' keyword. Otherwise you must use quotes if you wish to assign using the (.var:) syntax.   
   
Variables are dynamically reassignable using the same syntax


EXAMPLE   

    .var:myVar =  "someString"+"anotherOne" 
    .var:myVar =  808
    var myVar = /another/path