require 'sinatra'
require 'yaml'

get '/' do
  @title    = 'Insomnia 24/7 service status'
  @content  = YAML.load_file('status.yaml')
  @messages = YAML.load_file('messages.yaml')
  erb :index
end