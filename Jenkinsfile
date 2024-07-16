pipeline {
  agent any
  stages {
    stage('Checkout Code') {
      steps {
        git(url: 'https://github.com/liju-github/FoodBuddy-API', branch: 'master')
      }
    }

    stage('Check Environment Variable') {
      steps {
        script {
          def projectRoot = sh(script: 'echo $PROJECTROOT', returnStdout: true).trim()
          echo "PROJECTROOT: ${projectRoot}"
        }

      }
    }

    stage('Build') {
      steps {
        script {
          sh 'go build -o main'
        }

      }
    }

    stage('Run') {
      steps {
        script {
          sh 'nohup ./main &'
        }

      }
    }

  }
}