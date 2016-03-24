function cmd(id, data) {
    document.getElementById("body").className = "loading";
    var http = new XMLHttpRequest();
    http.open("POST", "/api/cmd/" + id, true);
    http.setRequestHeader("Connection", "close");

    http.onreadystatechange = function() {
        if (http.readyState == 4) {
            document.getElementById("body").className = "";
            if (http.status == 200) {
                popup(http.responseText)
            } else {
                popup("Server returned an error.")
            }
        }
    }
    
    http.send(data);
}

function logout() {
    document.getElementById("body").className = "loading";
    var http = new XMLHttpRequest();
    http.open("POST", "/api/logout", true);
    http.setRequestHeader("Connection", "close");

    http.onreadystatechange = function() {
        if (http.readyState == 4) {
            document.getElementById("body").className = "";
            if (http.status == 200) {
                document.location.replace("/")
            } else {
                popup("Server returned an error.")
            }
        }
    }
    
    http.send(null);
}
function popup(text) {
    document.getElementById("popup-content").innerHTML = text.replace("\n", "<br></br>");
    document.getElementById("popup").style = "opacity: 1; visibility: visible;";
}

function closepopup() {
    document.getElementById("popup").style = "opacity: 0; visibility: collapsed;";
}