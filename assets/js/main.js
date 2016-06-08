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
    $("#popup").openModal();
}

function closepopup() {
    $("#popup").closeModal();
}

function goToScheduler(id) {
    sessionStorage.setItem("pioneer-scheduler-cmd-id", id.toString());
    window.location.href = "/time";
}