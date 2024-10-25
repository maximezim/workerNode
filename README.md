# Variables d’Environnement

Ce projet utilise plusieurs variables d’environnement pour configurer la connexion MQTT, la communication avec le nœud maître et le système de traitement vidéo. Voici un aperçu des variables requises, leur rôle, et comment les configurer pour le nœud travailleur.

## Liste des Variables d’Environnement

1. **`MASTERNODE_URI`**  
   - **Rôle** : Adresse du nœud maître pour établir une connexion WebSocket. Cette connexion permet la réception des instructions et des données de traitement.
   - **Exemple de valeur** : `ws://localhost:8080`
   - **Définition** :
     ```bash
     export MASTERNODE_URI="ws://localhost:8080"
     ```

2. **`BROKER_URI`**  
   - **Rôle** : URI du broker MQTT utilisé pour la communication entre les nœuds et le traitement des données.
   - **Exemple de valeur** : `tcp://localhost:1883`
   - **Définition** :
     ```bash
     export BROKER_URI="tcp://localhost:1883"
     ```

3. **`MQTT_CLIENT_ID`**  
   - **Rôle** : Identifiant unique du client MQTT pour ce nœud. Utilisé pour l'authentification et la gestion des connexions.
   - **Exemple de valeur** : `worker-node-1`
   - **Définition** :
     ```bash
     export MQTT_CLIENT_ID="worker-node-1"
     ```

4. **`MQTT_USERNAME`**  
   - **Rôle** : Nom d’utilisateur pour l’authentification auprès du broker MQTT.
   - **Exemple de valeur** : `mqttUser`
   - **Définition** :
     ```bash
     export MQTT_USERNAME="mqttUser"
     ```

5. **`MQTT_PASSWORD`**  
   - **Rôle** : Mot de passe pour l’authentification auprès du broker MQTT.
   - **Exemple de valeur** : `securePass123`
   - **Définition** :
     ```bash
     export MQTT_PASSWORD="securePass123"
     ```

6. **`API_HOST`**  
   - **Rôle** : URI de l'API pour la création et la gestion des vidéos traitées. Utilisé pour vérifier l'existence d'une vidéo et pour créer de nouvelles entrées vidéo.
   - **Exemple de valeur** : `http://localhost:5000`
   - **Définition** :
     ```bash
     export API_HOST="http://localhost:5000"
     ```

## Comment Configurer les Variables

Pour définir ces variables d'environnement, utilisez la commande `export`, ou ajoutez-les au fichier `.env`. Exemple :

```bash
export MASTERNODE_URI="ws://localhost:8080"
export BROKER_URI="tcp://localhost:1883"
export MQTT_CLIENT_ID="worker-node-1"
export MQTT_USERNAME="mqttUser"
export MQTT_PASSWORD="securePass123"
export API_HOST="http://localhost:5000"
