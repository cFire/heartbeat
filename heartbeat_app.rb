require 'sinatra'
require 'yaml'

get '/' do
  @title    = 'Service status - Insomnia 24/7'
  @content  = YAML.load_file('status.yaml')
  @messages = YAML.load_file('messages.yaml')
  erb :index
end
