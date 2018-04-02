# frozen_string_literal: true

require 'sinatra'
require 'tty-table'
require 'yaml'

get '/' do
  @title    = 'Service status - Insomnia 24/7'
  @content  = YAML.load_file('status.yaml')
  @messages = YAML.load_file('messages.yaml')

  if request.user_agent.match?(/wget|curl/i)
    content_type :text
    erb :index_text
  else
    erb :index
  end
end
