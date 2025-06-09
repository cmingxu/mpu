API List
=====

## List Templates
### Get /api/templates

## Get Default Templates
### Get /api/templates/default

## List Movies
### List /api/movie

## Get Move
### Get /api/movie/:movie_id

## Create Movie
### Post /api/movies body: {"idea": "example idea"}

## Generate Script From Idea
###  Post /api/movies/:movie_id/generate_script body: {"movie_id": "example_movie_id", "idea": "example_idea", "prompt": "example idea"}

## Edit Script & Generate Corresponding English Subtitles
### Post /api/:movie_id/edit_script body: {"script": "example script"}

## Delete Script
### Post /api/:movie_id/:script_id/delete

## Generate Image for text clip
### Post /api/:movie_id/:script_id/generate_image


## List available bgm
### Post /api/bgms 

## Create BGM for whole movie
### Post /api/:movie_id/create_bgm body: {"bgm": "example_bgm"}

## Set transform filter for each clip
### Post /api/:movie_id/:script_id/set_filter body: {"filter": "example_filter"}

## Generate movie
### /api/:movie_id/generate
