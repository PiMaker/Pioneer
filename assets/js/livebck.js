var body = document.getElementsByTagName('body')[0];
var oldignore = 0;

function livebck() {
    ignore = new Date().getTime();
    body.style.backgroundImage = "url(/api/getbck?ignore=" + oldignore + "), url(/api/getbck?ignore=" + ignore + ")";
    oldignore = ignore;
    setTimeout(livebck, livebckinterval);
}

livebck();