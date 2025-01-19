// ****************** shared websocket code **************************

var wsUri = "ws://" + document.location.hostname + ":" + 8080 + "/chatSocket";
var websocket = new WebSocket(wsUri);

var username;
var output = document.getElementById("output");

websocket.onopen = function(evt) { onOpen(evt) };
websocket.onmessage = function(evt) { onMessage(evt) };
websocket.onerror = function(evt) { onError(evt) };

function sendMessage() {
    
    websocket.send("<img class='user-icon' src='/downloads/"+username+"/_user_icon.png' >" +username + ": " + textField.value);
}

function onOpen() {
    //writeToScreen("Connected to " + wsUri);
    //sendMessage(username + " joined");
}

function onMessage(evt) {
    writeToScreen(evt.data)
    /*
    if (evt.data.indexOf(pluginID) != -1){
        document.getElementById("chatFrame").contentWindow.proccessMessage(evt.data.substring(pluginID.length+1));
    }
    */
   /*
    const tokenArray = evt.data.split(":");
    var pluginID = tokenArray[0];
    console.log("pluginID:"+pluginID); 
    document.getElementById(pluginID+"Frame").contentWindow.proccessMessage(evt.data.substring(pluginID.length+1));
    */
    // this is an ugly workaround showing the need for plugin alias systems so different front ends can share functionality.
    /*
    if (pluginID == "chat"){
        document.getElementById("retroChatFrame").contentWindow.proccessMessage(evt.data.substring(pluginID.length+1));
    }*/
}

function onError(evt) {
    console.log("Websocket error:"+evt)
    //writeToScreen('<span style="color: red;">ERROR:</span> ' + evt.data);
}
/*******************************************************************/




function join(name) {
    //username = textField.value;
    username = name;
}
/*
function send_message() {
    parent.sendMessage(pluginID+":"+ username + ": " + textField.value);
}
    */

function proccessMessage(data){
    console.log("onMessage: " + data);
    if (data.indexOf("joined") != -1) {
        userField.innerHTML += data.substring(0, data.indexOf(" joined")) + "\n";
    } else {
        chatlogField.innerHTML += data + "\n";
    }
}

function writeToScreen(message) {
    output.innerHTML += message + "<br>";
    window.location = '#chatContainer';
    textField.value = "";
    textField.focus();
}


