var output = document.getElementById("output");
var server_ip = location.hostname
var socket = new WebSocket("ws://"+server_ip+":8080/stats");
var latestSpeed = 0;
var didset = false;
var chart = new Chart(document.getElementById("bw-chart"), {
  type: 'line',
  data: {
    labels: [  ],
    datasets: [
      { 
        data: [],
        label: "MBps",
        borderColor: "DarkGreen",
      }
    ]
  },
  options: {
    bezierCurve : false,
    title: {
      display: true,
      text: "Download Bandwidth"
    },
    scales: {
      xAxes: [{
        display: false
      }],
    }
  }
});
while (chart.data.labels.length < 60) {
  chart.data.labels.push("")
}

socket.onopen = function () {
    output.innerHTML = "Status: Connected\n";
};

socket.onclose = function () {
    output.innerHTML = "Status: Disconnected\n";
}

socket.onmessage = function (e) {
  var json = JSON.parse(e.data);

  if(!didset && json.Path != "") {
    document.getElementById("url").value = json.Path
    didset = true;
  }
  latestSpeed = json.BW / 1000;

  if (chart.data.datasets[0].data.length >= chart.data.labels.length) {
    chart.data.datasets[0].data.shift();
  }
  chart.data.datasets[0].data.push(latestSpeed);
  speed.innerHTML = "Download Speed: " + latestSpeed.toFixed(2) + " Mbps"
  chart.update();
};

function download() {
  var url = document.getElementById("url").value
  if(confirm("Download URL '" + url + "'")) {
    socket.send(url)
  }
}
