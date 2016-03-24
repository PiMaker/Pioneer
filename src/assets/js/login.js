function login() {
    var http = new XMLHttpRequest();
    http.withCredentials = true;
    http.open("POST", "/api/login", true);
    http.setRequestHeader("Connection", "close");

    http.onreadystatechange = function() {
        if (http.readyState == 4) {
            if (http.status == 200) {
                document.location.replace("/main");
            } else {
                document.getElementById("errorText").style.display = "block";
            }
        }
    }
    
    http.send(document.getElementById("passwordField").value);
}