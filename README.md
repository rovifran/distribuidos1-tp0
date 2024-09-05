# TP0: Docker + Comunicaciones + Concurrencia

Este repositorio implementa las resoluciones propuestas a los ejercicios del TP0 de la materia Sistemas Distribuidos 1 de la Facultad de Ingenieria, Universidad de Buenos Aires.  
  
A continuación se detallan las soluciones propuestas para cada ejercicio.

# Ejercicio 1
Para este ejercicio se pidio implementar un script que reciba por parametro la cantidad de clientes a generar, y el archivo `.yaml` a modificar para agregar los clientes.  
  
El script se puede correr con el siguiente comando:
```
./generar-compose.sh <archivo.yaml> <cantidad>
```
   
El algoritmo para generar el archivo esta implementado en Python y fue hecho mediante la modificacion y concatenacion de cadenas correspondientes al formato de un archivo `.yaml`.

# Ejercicio 2
En este ejercicio se tuvo que asociar los archivos de configuracion tanto para el cliente (`config.yaml`) como para el servidor (`config.ini`) para que estos no requieran levantar un nuevo container cada vez que se modificasen. Esto fue hecho con la creacion de un volumen y haciendo *bind mount* de los archivos de configuracion en el container.  
  
# Ejercicio 3
Para este ejercicio se tenia que implementar una forma de probar el echo server con la utilizacion de el programa `nc` (netcat). Las condiciones eran no exponer puertos, y no tener que correr `nc` de manera local, lo que significo tener que levantar un contenedor usando una imagen que fuese compatible con el comando `nc` y que este conectado a la misma red que esta conectado el server.
En mi caso en una primera instancia levante un container con la imagen `subfuzin/netcat` pero debido a un par de problemas con los tests y ejecucion misma del container, opte por una imagen de `alpine`, en el cual aplico el siguiente comando en una terminal:
```bash
echo $expected_response | nc -w 10 $server $server_port
```
Donde `expected_response` es la respuesta esperada del servidor, `server` es la direccion del servidor, y `server_port` es el puerto del servidor. Luego de correr este comando se compara la respuesta del servidor con la respuesta esperada, y en base a eso se determina si el test fue exitoso o no.

# Ejercicio 4
Para esta parte del codigo, se pedia implementar el manejo de la señal `SIGTERM` tanto para cliente como servidor, y hacer que estos terminen de forma *graceful*, cerrando archivos, sockets, etc. que haya quedado abierto.

## Cliente
Del lado del cliente hice lo siguiente:
1. Junto con la libreria `os/signal`, se creo un GracefulFinisher que lo que hace es ejecutarse en una go routine escuchando a traves de 2 channels en simultaneo: uno para la señal `SIGTERM` y otro para la señal de la funcion principal del programa para que termine su ejecucion y no quede esperando.  
2. Cuando se recibe la señal `SIGTERM`, se avisa a traves de otro channel, por el cual el cliente va a estar escuchando para saber que tiene que terminar su ejecucion. 
3. Se cierran los archivos y sockets que esten abiertos, y se termina la ejecucion del programa.

## Servidor
Del lado del servidor el manejo fue mas simple ya que no se necesita hacer esto que hice con las go routines:
1. Se asocia a la señal `SIGTERM` con un objeto llamado `SigTermSignalBinder`, que lo que hace es chequear la flag de finalizado el programa.
2. Siempre antes de que el servidor escuche por conexiones entrantes se chequea esta flag, asi como se tiene un manejador de excepciones para estos casos en donde se levanta la señal, para no terminar abruptamente.
3. Se cierra el socket del cliente y el socket listener del servidor.
  
La implementacion del graceful finish no varia mucho en el resto de los ejercicios pero siempre se tiene en cuenta el no dejar archivos o sockets abiertos.


# Ejercicio 5
En este ejercicio se requiere cambiar la funcionalidad del echo server que teniamos hasta ahora, para que funcionase como un servidor que recibe apuestas por parte de los clientes. Para esto habia que crear un protocolo de comunicacion entre el cliente y el servidor, en el cual se pueda pasar toda la informacion necesaria para realizar la apuesta.

## Protocolo  
La conexion se sigue haciendo por TCP entre el cliente y el servidor.  
  
La estructura que almacena la informacion del lado del cliente es la siguiente:
```go
type Bet struct {
    agency   uint8
    name     string
    surname  string
    dni      uint32
    birthday string
    number   uint16
}
```
Esto se hizo con la finalidad de optimizar el espacio ocupado por cada tipo de dato: por ejemplo el DNI se almacena en un `uint32` en vez de un `string` para ocupar 4 bytes en vez de 8.  
  
El formato con el que el cliente manda la informacion es el siguiente:
```
<cantidad_total_de_bytes_a_leer>
<agencia><longitud_nombre><nombre><longitud_apellido><apellido><dni><longitud_cumpleaños><cumpleaños><numero>
```
  
Estos campos no estan separados por ningun caracter en particular, sino que se envian uno a continuacion del otro. La longitud de este mensaje no es fija, pero es facilmente decodificable ya que para aquellos campos variables se envia la longitud del campo antes de enviar el campo en si. Para aquellos campos de longitud fija, no se envia la longitud, y son serializados en *little endian*.
  
El protocolo de desserializacion es implementado de forma inversa en el servidor: primero se lee la cantidad de bytes a leer, y luego se leen acordemente los bytes del mensaje hasta que se haya cumplido la cantidad de bytes a leer.  
  
La escritura y lectura de datos a traves de los sockets es tolerante a fallas del tipo *short read* y *short write* basandose en la cantidad de bytes que se esperan leer o escribir. En caso de que se escriban menos bytes de lo esperado, se sigue intentando escribir los bytes restantes hasta que se haya escrito la cantidad de bytes esperada. Lo mismo sucede con la lectura de datos.  
  
Se agrego funcionalidad a la clase `Bet` ya existente en el servidor, y se aprovecha esto para utilizar la funcion `store_bet`.

# Ejercicio 6
Para este ejercicio se pedia implementar la funcionalidad por el lado del cliente de mandar varias apuestas a la vez al servidor. Con el protocolo de comunicacion explicado anteriormente fue muy facil hacer las modificaciones requeridas para el cumplimiento del ejercicio.  
  
La cantidad maxima de apuestas que se pueden mandar esta determinado por el campo `batch.maxAmount` del archivo de configuracion del cliente. Este numero fue seteado en **134** con el siguiente criterio:  
Escaneando las apuestas de los archivos de las agencias (con el script `max_chars.sh`), se obtuvo que la apuesta mas larga tiene 61 caracteres. Entonces en el peor de los casos, la longitud de la apuesta seria de 61 bytes considerando la longitud de los campos variables y la longitud de la apuesta en bytes. Como la consigna pedia no sobrepasar los 8kb de datos, se hizo el calculo de la cantidad maxima de apuestas que se pueden mandar con este formato, y resulto en el numero **134**.  
Aunque la cantidad total de apuestas puede llegar a ser mas que ese maximo calculado (gracias a que la longitud de las apuestas no es fija), se decidio respetar el campo `batch.maxAmount`.  
  
La estructura de los mensajes que manda el cliente es la siguiente:
```
<tamaño en bytes de apuestas totales>
<tamaño en bytes de apuesta 1>
<apuesta1>
<tamaño en bytes de apuesta 2>
<apuesta2>
...
<tamaño en bytes de apuesta N>
<apuestaN>
```
Mantuve la serializacion de la apuesta igual, lo unico que se modifico fue que el campo inicial ahora es la cantidad total de bytes a leer, y luego se sigue con la serializacion de las apuestas.  
  
Por el lado del servidor, tambien se hizo la modificacion correspondiente para que pueda leer la cantidad de bytes de apuestas totales, y luego leer las apuestas correspondientes. Luego se sigue con la serializacion de todas las apuestas (aca se asume que no hay errores gracias a la tolerancia a fallas implementada en el protocolo de comunicacion y tambien al mecanismo de integridad de los datos que ofrece TCP), su almacenamiento en el archivo de apuestas, y la respuesta al cliente con la cantidad de apuestas recibidas. Esta parte del protocolo fue la que sufrio cambios con respecto a la implementacion anterior: ahora se manda la cantidad de apuestas procesadas o un -1 indicando que hubo un error.  
