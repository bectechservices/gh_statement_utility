function toggleVisible() {
    var temp = document.getElementById("auth[passwordhasVisibility]");
    if (temp.type === "password") {
      temp.type = "text";
    } else {
      temp.type = "password";
    }
  }