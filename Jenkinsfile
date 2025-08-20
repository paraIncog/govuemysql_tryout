pipeline {
  agent any

  options { timestamps() }

  // ✅ names unquoted; values are simple strings
  environment {
    REGISTRY         = 'docker.io'            // or your private registry
    PROJECT          = 'paraIncog/govue'      // change to what you push under
    DOCKER_CREDS     = 'docker-reg-cred'      // Jenkins credentialsId
    DEPLOY_HOST      = 'ubuntu@your-server'   // if using remote deploy
    DEPLOY_SSH_CRED  = 'deploy-ssh-key'       // Jenkins credentialsId
    COMPOSE_DIR      = '/opt/myapp/deploy'    // remote compose dir
    IMAGE_TAG        = "bootstrap-${BUILD_NUMBER}" // simple initial (valid here)
  }

  stages {
    stage('Checkout') {
      steps {
        git branch: 'main', url: 'https://github.com/paraIncog/govuemysql_tryout.git'
      }
    }

    // Compute dynamic variables AFTER checkout
    stage('Init env') {
      steps {
        script {
          env.GIT_SHA     = sh(script: 'git rev-parse --short=7 HEAD', returnStdout: true).trim()
          env.BRANCH_NAME = env.BRANCH_NAME ?: sh(script: 'git rev-parse --abbrev-ref HEAD', returnStdout: true).trim()
          env.IMAGE_TAG   = (env.BRANCH_NAME == 'main' || env.BRANCH_NAME.startsWith('release/')) ?
                            "prod-${env.GIT_SHA}" : "dev-${env.GIT_SHA}"
          echo "BRANCH=${env.BRANCH_NAME}, SHA=${env.GIT_SHA}, TAG=${env.IMAGE_TAG}"
        }
      }
    }

    // --- Backend: test & build image ---
    stage('Backend: Test') {
      steps {
        dir('backend') {
          sh 'go version || true'
          sh 'go test ./... || true'   // make strict by removing "|| true"
        }
      }
    }
    stage('Backend: Build Image') {
      steps {
        dir('backend') {
          sh '''
            docker build \
              --label org.opencontainers.image.revision=$GIT_SHA \
              --label org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
              -t $REGISTRY/$PROJECT/backend:$IMAGE_TAG .
          '''
        }
      }
    }

    // --- Frontend: build & image ---
    stage('Frontend: Build Image') {
      steps {
        dir('frontend') {
          sh '''
            npm ci
            npm run test --if-present
            npm run build --if-present
            docker build \
              --label org.opencontainers.image.revision=$GIT_SHA \
              --label org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
              -t $REGISTRY/$PROJECT/frontend:$IMAGE_TAG .
          '''
        }
      }
    }

    // --- Push images ---
    stage('Push Images') {
      steps {
        withCredentials([usernamePassword(credentialsId: env.DOCKER_CREDS, usernameVariable: 'USER', passwordVariable: 'PASS')]) {
          sh 'echo $PASS | docker login $REGISTRY -u $USER --password-stdin'
        }
        sh '''
          docker push $REGISTRY/$PROJECT/backend:$IMAGE_TAG
          docker push $REGISTRY/$PROJECT/frontend:$IMAGE_TAG
        '''
      }
    }

    // --- Deploy (pick ONE of the two) ---

    // A) Same host as Jenkins
    stage('Deploy (local compose)') {
      when { anyOf { branch 'main'; expression { env.BRANCH_NAME?.startsWith('release/') } } }
      steps {
        sh '''
          export IMAGE_TAG=$IMAGE_TAG REGISTRY=$REGISTRY PROJECT=$PROJECT
          cd deploy
          docker compose pull
          docker compose up -d --remove-orphans
          docker compose ps
        '''
      }
    }

    // B) Remote host via SSH
    // Comment out A) and enable this if you deploy to another server.
    /*
    stage('Deploy (remote compose)') {
      when { anyOf { branch 'main'; expression { env.BRANCH_NAME?.startsWith('release/') } } }
      steps {
        sshagent(credentials: [env.DEPLOY_SSH_CRED]) {
          sh """
            ssh -o StrictHostKeyChecking=no ${env.DEPLOY_HOST} \\
              'export IMAGE_TAG=${IMAGE_TAG} REGISTRY=${REGISTRY} PROJECT=${PROJECT} && \\
               cd ${COMPOSE_DIR} && docker compose pull && docker compose up -d --remove-orphans && docker compose ps'
          """
        }
      }
    }
    */
  }

  post {
    always  { sh 'docker logout $REGISTRY || true' }
    success { echo "Deployed tag ${env.IMAGE_TAG}" }
    failure { echo "Build or deploy failed." }
  }
}
