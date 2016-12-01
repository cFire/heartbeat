class ReadTimeout < Exception
end

class CheckService
  require 'timeout'

  def initialize(server, service, address, parameters)
    @server     = server
    @service    = service
    @address    = address
    @parameters = parameters
    @status
    @responseTime
    @checkTime
  end

  def server
    @server
  end

  def service
    @service
  end

  def status
    @status
  end

  def responseTime
    @responseTime
  end

  def checkTime
    @checkTime
  end

  def run
    @status       = 'DOWN'
    @responseTime = 0
    @checkTime    = Time.now
  end
end

class CheckHTTP < CheckService
  require 'nokogiri'
  require 'open-uri'
  require 'openssl'

  def run
    response = ''
    start    = Time.now.to_f * 1000

    begin
      Timeout::timeout(10) {
        uri = ''
        if(@parameters['ssl'])
          uri = 'https://'
        else
          uri = 'http://'
        end
        if(@parameters.has_key?('vhost'))
          uri += @parameters['vhost']  
        else
          uri += @address
        end
        uri += ":#{@parameters['port']}#{@parameters['get']}"

        doc = Nokogiri::HTML open(uri, redirect: false)
        if(doc.to_s.include?(@parameters['expect']))
          @status = 'OK'
        else
          @status = 'WARNING'
        end
      }
    rescue OpenURI::HTTPRedirect => e
      if(e.to_s.include?(@parameters['expect']))
        @status = 'OK'
      else
        @status = 'WARNING'
      end
    rescue OpenURI::HTTPError => e
      if(e.to_s.include?(@parameters['expect']))
        @status = 'OK'
      else
        @status = 'WARNING'
      end
    rescue Timeout::Error
      @status = 'DOWN'
    rescue
      @status = 'DOWN'
    end

    @responseTime = (Time.now.to_f * 1000 - start).floor
    @checkTime    = Time.now
  end
end

class CheckTCP < CheckService
  require 'socket'
  require 'openssl'

  def run
    socket   = TCPSocket
    response = ''
    start    = Time.now.to_f * 1000

    begin
      Timeout::timeout(10) {
        s = TCPSocket.open(@address, @parameters['port'])
        if(@parameters['ssl'])
          ssl_context = OpenSSL::SSL::SSLContext.new()
          socket = OpenSSL::SSL::SSLSocket.new(s, ssl_context)
          socket.sync_close = true
          socket.connect
        else
          socket = s
        end
        

        if(@parameters.has_key?('send'))
          socket.puts(@parameters['send'])
        end

        Timeout::timeout(2, ReadTimeout) {
          while line = socket.gets
            response += line
            if response.include?(@parameters['expect'])
              @status = 'OK'
              socket.close
              break
            end
          end
        }
      }
    rescue ReadTimeout
      if response.include?(@parameters['expect'])
        @status = 'OK'
      else
        @status = 'WARNING'
      end
      socket.close
    rescue Timeout::Error
      @status = 'DOWN'
    rescue
      @status = 'DOWN'
    end

    @responseTime = (Time.now.to_f * 1000 - start).floor
    @checkTime    = Time.now
  end
end
