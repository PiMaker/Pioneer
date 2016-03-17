function cmd(id, data) {
    var http = new XMLHttpRequest();
    http.open("POST", "/api/cmd/" + id, true);
    http.setRequestHeader("Connection", "close");

    http.onreadystatechange = function() {
        if (http.readyState == 4) {
            if (http.status == 200) {
                alert("Successful!")
            } else {
                alert("Server returned an error.")
            }
        }
    }
    
    http.send(data);
}

function logout() {
    var http = new XMLHttpRequest();
    http.open("POST", "/api/logout", true);
    http.setRequestHeader("Connection", "close");

    http.onreadystatechange = function() {
        if (http.readyState == 4) {
            if (http.status == 200) {
                document.location.assign("/")
            } else {
                alert("Server returned an error.")
            }
        }
    }
    
    http.send(null);
}