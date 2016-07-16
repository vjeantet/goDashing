<?php

$arr = array(
				"labels"=> array("January", "February", "March", "April", "May", "June", "July"),
				"datasets"=> array(
					array(
						"label"=>           "My First dataset",
						"fillColor"=>       "rgba(220,220,220,0.5)",
						"strokeColor"=>     "rgba(220,220,220,0.8)",
						"highlightFill"=>   "rgba(220,220,220,0.75)",
						"highlightStroke"=> "rgba(220,220,220,1)",
						"data"=>            array(rand(1,60), rand(1,42), rand(1,82), rand(1,13), rand(1,57), 5, 57),
					),
					array(
						"label"=>           "My Second dataset",
						"fillColor"=>       "rgba(151,187,205,0.5)",
						"strokeColor"=>     "rgba(151,187,205,0.8)",
						"highlightFill"=>   "rgba(151,187,205,0.75)",
						"highlightStroke"=> "rgba(151,187,205,1)",
						"data"=>            array(60, rand(1,80), 62, rand(1,63), 67, rand(1,50), 57),
					)
				),
				"options"=> array("scaleFontColor"=> "#fff"),
	);

echo json_encode($arr);