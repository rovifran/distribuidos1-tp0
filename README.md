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
  
---
el resto de secciones las hago despues 
---

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
