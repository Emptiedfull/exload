




    if (socket){
        socket.close()
    }

    var socket = new WebSocket("/admin/ws/pen_graph/"+document.querySelector(".active").innerHTML)

    var ctx = document.getElementById('penChart').getContext('2d');
    var myChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [], 
            datasets: [
                
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
           
            scales: {
                x: {
                    type: 'time', 
                    time: {
                        unit: 'second', 
                        stepsize: 5,
                        tooltipFormat: 'PPpp', 
                        displayFormats: {
                            second: 'HH:mm:ss' 
                        }
                    },
                    
                    ticks: {
                        maxTicksLimit: 5,
                        source: 'auto', 
                        autoSkip: false,
                        maxRotation: 0, 
                        color: 'rgba(255, 255, 255, 0.40)'
                    }
                },
                y: {
                    beginAtZero: true,
                   
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.40)',
                        maxTicksLimit: 8
                    },
                    suggestedMin: 50, 
                    suggestedMax: 300 
                }
                
            },
            plugins: {
                legend: {
                    position: 'right', 
                    labels: {
                        color: '#FFFFFF'
                    }
                }
            }
        }
    });

   

   var colourList = ["#9F9FF8","#92BFFF","#94E9B8"]

    
    

   
    socket.onmessage = function(event) {
        let sample = JSON.parse(event.data);
        
        data = {
            pens:sample
        }

        handleChartUpdate(data)

        
    };

    // const sample = {
    //     timestamp: new Date().toISOString(),
    //     pens : {
    //         "/api":Math.random()* 100,
    //         "/admin": Math.random()
    //     }
    // }

    function handleChartUpdate(data){
        let timestamp = new Date().toISOString()

        for (let key in data.pens){
            if (data.pens.hasOwnProperty(key)) {
                let value = data.pens[key];
                console.log(`Key: ${key}, Value: ${value}`);
               
                let dataset = myChart.data.datasets.find(ds => ds.label === key)

                color = colourList.pop()

                if (!dataset){
                    dataset =  {
                        label:key,
                        data: [], 
                        backgroundColor: color,
                        borderColor:color, 
                        borderWidth: 1,
                        fill: false,
                        yAxisID: 'y'
                    }
                    myChart.data.datasets.push(dataset);
                }

                dataset.data.push({ x: timestamp, y: value });

                
                if (dataset.data.length > 8) {
                    dataset.data.shift();
                }

                
                if (!myChart.data.labels.includes(timestamp)) {
                    myChart.data.labels.push(timestamp);

                    
                    if (myChart.data.labels.length > 8) {
                        myChart.data.labels.shift();
                    }
                }

                if (!myChart.data.labels.includes(timestamp)) {
                    myChart.data.labels.push(timestamp);
                }
               
            }
        }

        myChart.update("none");
    }

    function getRandomElement(list) {
        return list[Math.floor(Math.random() * list.length)];
    }
   

