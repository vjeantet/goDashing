<?php
$arr = array(
		"segments"=> array(
			array(
				"value" =>    rand(1,20),
				"color"=>     "#F7464A",
				"highlight"=> "#FF5A5E",
				"label"=>     "January",
			),	
			array(
				"value"=>     rand(1,30),
				"color"=>     "#46BFBD",
				"highlight"=> "#5AD3D1",
				"label"=>     "February",
			), 
			array(
				"value"=>     rand(1,30),
				"color"=>     "#FDB45C",
				"highlight"=> "#FFC870",
				"label"=>     "March",
			), 
			array(
				"value"=>     rand(1,30),
				"color"=>     "#949FB1",
				"highlight"=> "#A8B3C5",
				"label"=>     "April",
			), 
			array(
				"value"=>     rand(1,30),
				"color"=>     "#4D5360",
				"highlight"=> "#4D5360",
				"label"=>     "April",
			),
		),
		"options"=> array("segmentStrokeColor"=> "#333"),
	);

echo json_encode($arr);